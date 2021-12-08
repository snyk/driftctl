package pkg

import (
	"fmt"
	"time"

	"github.com/jmespath/go-jmespath"
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/pkg/memstore"
	globaloutput "github.com/snyk/driftctl/pkg/output"

	"github.com/snyk/driftctl/pkg/alerter"
	"github.com/snyk/driftctl/pkg/analyser"
	"github.com/snyk/driftctl/pkg/cmd/scan/output"
	"github.com/snyk/driftctl/pkg/filter"
	"github.com/snyk/driftctl/pkg/iac/config"
	"github.com/snyk/driftctl/pkg/iac/terraform/state/backend"
	"github.com/snyk/driftctl/pkg/middlewares"
	"github.com/snyk/driftctl/pkg/resource"
)

type ScanOptions struct {
	Coverage         bool
	Detect           bool
	From             []config.SupplierConfig
	To               string
	Output           []output.OutputConfig
	Filter           *jmespath.JMESPath
	Quiet            bool
	BackendOptions   *backend.Options
	StrictMode       bool
	DisableTelemetry bool
	ProviderVersion  string
	ConfigDir        string
	DriftignorePath  string
	Deep             bool
}

type DriftCTL struct {
	remoteSupplier           resource.RemoteSupplier
	iacSupplier              resource.Supplier
	alerter                  alerter.AlerterInterface
	analyzer                 *analyser.Analyzer
	resourceFactory          resource.ResourceFactory
	scanProgress             globaloutput.Progress
	iacProgress              globaloutput.Progress
	resourceSchemaRepository resource.SchemaRepositoryInterface
	opts                     *ScanOptions
	store                    memstore.Store
}

func NewDriftCTL(remoteSupplier resource.RemoteSupplier,
	iacSupplier resource.Supplier,
	alerter *alerter.Alerter,
	analyzer *analyser.Analyzer,
	resFactory resource.ResourceFactory,
	opts *ScanOptions,
	scanProgress globaloutput.Progress,
	iacProgress globaloutput.Progress,
	resourceSchemaRepository resource.SchemaRepositoryInterface,
	store memstore.Store) *DriftCTL {
	return &DriftCTL{
		remoteSupplier,
		iacSupplier,
		alerter,
		analyzer,
		resFactory,
		scanProgress,
		iacProgress,
		resourceSchemaRepository,
		opts,
		store,
	}
}

func (d DriftCTL) Run() (*analyser.Analysis, error) {
	start := time.Now()
	remoteResources, resourcesFromState, err := d.enumerateResources()
	if err != nil {
		return nil, err
	}

	middleware := middlewares.NewChain(
		middlewares.NewRoute53RecordIDReconcilier(),
		middlewares.NewRoute53DefaultZoneRecordSanitizer(),
		middlewares.NewS3BucketAcl(),
		middlewares.NewAwsInstanceBlockDeviceResourceMapper(d.resourceFactory),
		middlewares.NewAwsDefaultSecurityGroupRule(),
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
		middlewares.NewAwsDefaultNetworkACL(),
		middlewares.NewAwsDefaultNetworkACLRule(),
		middlewares.NewAwsNetworkACLExpander(d.resourceFactory),
		middlewares.NewAwsNatGatewayEipAssoc(),
		middlewares.NewAwsBucketPolicyExpander(d.resourceFactory),
		middlewares.NewAwsSQSQueuePolicyExpander(d.resourceFactory, d.resourceSchemaRepository),
		middlewares.NewAwsDefaultSQSQueuePolicy(),
		middlewares.NewAwsSNSTopicPolicyExpander(d.resourceFactory, d.resourceSchemaRepository),
		middlewares.NewAwsRoleManagedPolicyExpander(d.resourceFactory),
		middlewares.NewTagsAllManager(),
		middlewares.NewEipAssociationExpander(d.resourceFactory),
		middlewares.NewRDSClusterInstanceExpander(d.resourceFactory),
		middlewares.NewAwsApiGatewayDeploymentExpander(d.resourceFactory),
		middlewares.NewAwsApiGatewayResourceExpander(d.resourceFactory),
		middlewares.NewAwsApiGatewayRestApiExpander(d.resourceFactory),
		middlewares.NewAwsApiGatewayRestApiPolicyExpander(d.resourceFactory),
		middlewares.NewAwsConsoleApiGatewayGatewayResponse(),

		middlewares.NewGoogleIAMBindingTransformer(d.resourceFactory),
		middlewares.NewGoogleIAMPolicyTransformer(d.resourceFactory),

		middlewares.NewAzurermRouteExpander(d.resourceFactory),
		middlewares.NewAzurermSubnetExpander(d.resourceFactory),
	)

	if !d.opts.StrictMode {
		middleware = append(middleware,
			middlewares.NewAwsDefaults(),
			middlewares.NewGoogleLegacyBucketIAMMember(),
			middlewares.NewGoogleDefaultIAMMember(),
		)
	}

	logrus.Debug("Ready to run middlewares")
	err = middleware.Execute(&remoteResources, &resourcesFromState)
	if err != nil {
		return nil, err
	}

	if d.opts.Filter != nil {
		engine := filter.NewFilterEngine(d.opts.Filter)
		remoteResources, err = engine.Run(remoteResources)
		if err != nil {
			return nil, err
		}
		resourcesFromState, err = engine.Run(resourcesFromState)
		if err != nil {
			return nil, err
		}
	}

	analysis := analyser.NewAnalysis(analyser.AnalyzerOptions{Deep: d.opts.Deep})
	analysis = d.analyzer.CompareEnumeration(analysis, remoteResources, resourcesFromState)
	if err != nil {
		return nil, err
	}

	managedResources, err := d.readResources(analysis.Managed())
	if err != nil {
		return nil, err
	}

	analysis = d.analyzer.CompleteAnalysis(analysis, managedResources, resourcesFromState)

	// Sort resources by Terraform Id
	// The purpose is to have a predictable output
	analysis.SortResources()

	analysis.Duration = time.Since(start)
	analysis.Date = time.Now()

	d.store.Bucket(memstore.TelemetryBucket).Set("total_resources", analysis.Summary().TotalResources)
	d.store.Bucket(memstore.TelemetryBucket).Set("total_managed", analysis.Summary().TotalManaged)
	d.store.Bucket(memstore.TelemetryBucket).Set("duration", uint(analysis.Duration.Seconds()+0.5))

	return analysis, nil
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

func (d DriftCTL) enumerateResources() (remoteResources []*resource.Resource, resourcesFromState []*resource.Resource, err error) {
	logrus.Info("Start reading IaC")
	d.iacProgress.Start()
	resourcesFromState, err = d.iacSupplier.Resources()
	d.iacProgress.Stop()
	if err != nil {
		return nil, nil, err
	}

	logrus.Info("Start enumerating cloud provider resources")
	d.scanProgress.Start()
	defer d.scanProgress.Stop()
	remoteResources, err = d.remoteSupplier.EnumerateResources()
	if err != nil {
		return nil, nil, err
	}

	return remoteResources, resourcesFromState, err
}

func (d DriftCTL) readResources(managedResources []*resource.Resource) ([]*resource.Resource, error) {
	logrus.WithField("count", len(managedResources)).Info("Fetching details of managed resources")
	d.scanProgress.Start()
	defer d.scanProgress.Stop()
	return d.remoteSupplier.ReadResources(managedResources)
}
