package pkg

import (
	"fmt"
	"time"

	globaloutput "github.com/cloudskiff/driftctl/pkg/output"
	"github.com/jmespath/go-jmespath"
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/cloudskiff/driftctl/pkg/cmd/scan/output"
	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/iac/config"
	"github.com/cloudskiff/driftctl/pkg/iac/terraform/state/backend"
	"github.com/cloudskiff/driftctl/pkg/middlewares"
	"github.com/cloudskiff/driftctl/pkg/resource"
)

type ScanOptions struct {
	Coverage         bool
	Detect           bool
	From             []config.SupplierConfig
	To               string
	Output           output.OutputConfig
	Filter           *jmespath.JMESPath
	Quiet            bool
	BackendOptions   *backend.Options
	StrictMode       bool
	DisableTelemetry bool
	ProviderVersion  string
}

type DriftCTL struct {
	remoteSupplier           resource.Supplier
	iacSupplier              resource.Supplier
	alerter                  alerter.AlerterInterface
	analyzer                 analyser.Analyzer
	filter                   *jmespath.JMESPath
	resourceFactory          resource.ResourceFactory
	strictMode               bool
	scanProgress             globaloutput.Progress
	iacProgress              globaloutput.Progress
	resourceSchemaRepository resource.SchemaRepositoryInterface
}

func NewDriftCTL(remoteSupplier resource.Supplier,
	iacSupplier resource.Supplier,
	alerter *alerter.Alerter,
	resFactory resource.ResourceFactory,
	opts *ScanOptions,
	scanProgress globaloutput.Progress,
	iacProgress globaloutput.Progress,
	resourceSchemaRepository resource.SchemaRepositoryInterface) *DriftCTL {
	return &DriftCTL{
		remoteSupplier,
		iacSupplier,
		alerter,
		analyser.NewAnalyzer(alerter),
		opts.Filter,
		resFactory,
		opts.StrictMode,
		scanProgress,
		iacProgress,
		resourceSchemaRepository,
	}
}

func (d DriftCTL) Run() (*analyser.Analysis, error) {
	start := time.Now()
	remoteResources, resourcesFromState, err := d.scan()
	if err != nil {
		return nil, err
	}

	middleware := middlewares.NewChain(
		middlewares.NewRoute53RecordIDReconcilier(),
		middlewares.NewRoute53DefaultZoneRecordSanitizer(),
		middlewares.NewS3BucketAcl(),
		middlewares.NewAwsInstanceBlockDeviceResourceMapper(d.resourceFactory),
		middlewares.NewVPCDefaultSecurityGroupSanitizer(),
		middlewares.NewVPCSecurityGroupRuleSanitizer(d.resourceFactory),
		middlewares.NewIamPolicyAttachmentTransformer(d.resourceFactory),
		middlewares.NewIamPolicyAttachmentExpander(d.resourceFactory),
		middlewares.AwsInstanceEIP{},
		middlewares.NewAwsDefaultInternetGatewayRoute(),
		middlewares.NewAwsDefaultInternetGateway(),
		middlewares.NewAwsDefaultVPC(),
		middlewares.NewAwsDefaultSubnet(),
		middlewares.NewAwsRouteTableExpander(d.alerter, d.resourceFactory),
		middlewares.NewAwsDefaultRouteTable(),
		middlewares.NewAwsDefaultRoute(),
		middlewares.NewAwsNatGatewayEipAssoc(),
		middlewares.NewAwsBucketPolicyExpander(d.resourceFactory),
		middlewares.NewAwsSqsQueuePolicyExpander(d.resourceFactory, d.resourceSchemaRepository),
		middlewares.NewAwsDefaultSqsQueuePolicy(),
		middlewares.NewAwsSNSTopicPolicyExpander(d.resourceFactory, d.resourceSchemaRepository),
	)

	if !d.strictMode {
		middleware = append(middleware,
			middlewares.NewAwsDefaults(),
		)
	}

	logrus.Debug("Ready to run middlewares")
	err = middleware.Execute(&remoteResources, &resourcesFromState)
	if err != nil {
		return nil, err
	}

	if d.filter != nil {
		engine := filter.NewFilterEngine(d.filter)
		remoteResources, err = engine.Run(remoteResources)
		if err != nil {
			return nil, err
		}
		resourcesFromState, err = engine.Run(resourcesFromState)
		if err != nil {
			return nil, err
		}
	}

	logrus.Debug("Checking for driftignore")
	driftIgnore := filter.NewDriftIgnore()

	analysis, err := d.analyzer.Analyze(remoteResources, resourcesFromState, driftIgnore)
	analysis.Duration = time.Since(start)

	if err != nil {
		return nil, err
	}

	return &analysis, nil
}

func (d DriftCTL) Stop() {
	stoppableSupplier, ok := d.remoteSupplier.(resource.StoppableSupplier)
	if ok {
		logrus.WithFields(logrus.Fields{
			"supplier": fmt.Sprintf("%T", d.remoteSupplier),
		}).Debug("Stopping remote supplier")
		stoppableSupplier.Stop()
	}

	stoppableSupplier, ok = d.iacSupplier.(resource.StoppableSupplier)
	if ok {
		stoppableSupplier.Stop()
	}
}

func (d DriftCTL) scan() (remoteResources []resource.Resource, resourcesFromState []resource.Resource, err error) {
	logrus.Info("Start reading IaC")
	d.iacProgress.Start()
	resourcesFromState, err = d.iacSupplier.Resources()
	d.iacProgress.Stop()
	if err != nil {
		return nil, nil, err
	}

	logrus.Info("Start scanning cloud provider")
	d.scanProgress.Start()
	defer d.scanProgress.Stop()
	remoteResources, err = d.remoteSupplier.Resources()
	if err != nil {
		return nil, nil, err
	}

	return remoteResources, resourcesFromState, err
}
