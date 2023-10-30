package pkg

import (
	"fmt"
	"time"

	"github.com/jmespath/go-jmespath"
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/analyser"
	"github.com/snyk/driftctl/pkg/cmd/scan/output"
	"github.com/snyk/driftctl/pkg/filter"
	"github.com/snyk/driftctl/pkg/iac/config"
	"github.com/snyk/driftctl/pkg/iac/terraform/state/backend"
	"github.com/snyk/driftctl/pkg/memstore"
	"github.com/snyk/driftctl/pkg/middlewares"
	globaloutput "github.com/snyk/driftctl/pkg/output"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

type FmtOptions struct {
	Output output.OutputConfig
}

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
	Driftignores     []string
}

type DriftCTL struct {
	remoteSupplier           resource.Supplier
	iacSupplier              dctlresource.IaCSupplier
	alerter                  alerter.AlerterInterface
	analyzer                 *analyser.Analyzer
	resourceFactory          resource.ResourceFactory
	scanProgress             globaloutput.Progress
	iacProgress              globaloutput.Progress
	resourceSchemaRepository dctlresource.SchemaRepositoryInterface
	opts                     *ScanOptions
	store                    memstore.Store
}

func NewDriftCTL(remoteSupplier resource.Supplier,
	iacSupplier dctlresource.IaCSupplier,
	alerter *alerter.Alerter,
	analyzer *analyser.Analyzer,
	resFactory resource.ResourceFactory,
	opts *ScanOptions,
	scanProgress globaloutput.Progress,
	iacProgress globaloutput.Progress,
	resourceSchemaRepository dctlresource.SchemaRepositoryInterface,
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
	remoteResources, resourcesFromState, err := d.scan()
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
		middlewares.NewAwsBucketPolicyExpander(d.resourceFactory),
		middlewares.NewAwsSQSQueuePolicyExpander(d.resourceFactory, d.resourceSchemaRepository),
		middlewares.NewAwsDefaultSQSQueuePolicy(),
		middlewares.NewAwsSNSTopicPolicyExpander(d.resourceFactory, d.resourceSchemaRepository),
		middlewares.NewAwsRoleManagedPolicyExpander(d.resourceFactory),
		middlewares.NewTagsAllManager(),
		middlewares.NewEipAssociationExpander(d.resourceFactory),
		middlewares.NewAwsNatGatewayEipAssoc(),
		middlewares.NewRDSClusterInstanceExpander(d.resourceFactory),
		middlewares.NewAwsApiGatewayDeploymentExpander(d.resourceFactory),
		middlewares.NewAwsApiGatewayResourceExpander(d.resourceFactory),
		middlewares.NewAwsApiGatewayApiExpander(d.resourceFactory),
		middlewares.NewAwsApiGatewayRestApiPolicyExpander(d.resourceFactory),
		middlewares.NewAwsConsoleApiGatewayGatewayResponse(),
		middlewares.NewAwsApiGatewayDomainNamesReconciler(),
		middlewares.NewAwsApiGatewayBasePathMappingReconciler(),
		middlewares.NewAwsEbsEncryptionByDefaultReconciler(d.resourceFactory),
		middlewares.NewAwsALBTransformer(d.resourceFactory),
		middlewares.NewAwsALBListenerTransformer(d.resourceFactory),

		middlewares.NewGoogleIAMBindingTransformer(d.resourceFactory),
		middlewares.NewGoogleIAMPolicyTransformer(d.resourceFactory),
		middlewares.NewGoogleComputeInstanceGroupManagerReconciler(),

		middlewares.NewAzurermRouteExpander(d.resourceFactory),
		middlewares.NewAzurermSubnetExpander(d.resourceFactory),
		middlewares.NewAwsS3BucketPublicAccessBlockReconciler(),
	)

	if !d.opts.StrictMode {
		middleware = append(middleware,
			middlewares.NewAwsDefaults(),
			middlewares.NewGoogleLegacyBucketIAMMember(),
			middlewares.NewGoogleDefaultIAMMember(),
			middlewares.NewAwsDefaultApiGatewayAccount(),
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

	analysis, err := d.analyzer.Analyze(remoteResources, resourcesFromState)
	if err != nil {
		return nil, err
	}

	analysis.SetIaCSourceCount(d.iacSupplier.SourceCount())
	analysis.Duration = time.Since(start)
	analysis.Date = time.Now()

	d.store.Bucket(memstore.TelemetryBucket).Set("total_resources", analysis.Summary().TotalResources)
	d.store.Bucket(memstore.TelemetryBucket).Set("total_managed", analysis.Summary().TotalManaged)
	d.store.Bucket(memstore.TelemetryBucket).Set("duration", uint(analysis.Duration.Seconds()+0.5))
	d.store.Bucket(memstore.TelemetryBucket).Set("iac_source_count", d.iacSupplier.SourceCount())

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

func (d DriftCTL) scan() (remoteResources []*resource.Resource, resourcesFromState []*resource.Resource, err error) {
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

	// We do a normalization pass to resources from remote because resource in IaC supplier
	// are already created using DriftctlFactory.CreateAbstractResource and thus are already normalized
	var normalizedRemoteResources []*resource.Resource
	for _, res := range remoteResources {
		attrs := resource.Attributes{}
		if res.Attributes() != nil {
			attrs = *res.Attributes()
		}
		normalizedRes := d.resourceFactory.CreateAbstractResource(res.ResourceType(), res.ResourceId(), attrs)
		normalizedRemoteResources = append(normalizedRemoteResources, normalizedRes)
	}

	return normalizedRemoteResources, resourcesFromState, err
}
