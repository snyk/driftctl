package pkg

import (
	"fmt"

	"github.com/jmespath/go-jmespath"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/middlewares"
	"github.com/cloudskiff/driftctl/pkg/resource"
)

type DriftCTL struct {
	remoteSupplier  resource.Supplier
	iacSupplier     resource.Supplier
	alerter         alerter.AlerterInterface
	analyzer        analyser.Analyzer
	filter          *jmespath.JMESPath
	resourceFactory resource.ResourceFactory
	strictMode      bool
}

func NewDriftCTL(remoteSupplier resource.Supplier, iacSupplier resource.Supplier, filter *jmespath.JMESPath, alerter *alerter.Alerter, resFactory resource.ResourceFactory, strictMode bool) *DriftCTL {
	return &DriftCTL{remoteSupplier, iacSupplier, alerter, analyser.NewAnalyzer(alerter), filter, resFactory, strictMode}
}

func (d DriftCTL) Run() (*analyser.Analysis, error) {
	remoteResources, resourcesFromState, err := d.scan()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	middleware := middlewares.NewChain(
		middlewares.NewRoute53DefaultZoneRecordSanitizer(),
		middlewares.NewS3BucketAcl(),
		middlewares.NewAwsInstanceBlockDeviceResourceMapper(d.resourceFactory),
		middlewares.NewVPCDefaultSecurityGroupSanitizer(),
		middlewares.NewVPCSecurityGroupRuleSanitizer(d.resourceFactory),
		middlewares.NewIamPolicyAttachmentSanitizer(),
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
		middlewares.NewAwsSqsQueuePolicyExpander(d.resourceFactory),
		middlewares.NewAwsDefaultSqsQueuePolicy(),
		middlewares.NewAwsSNSTopicPolicyExpander(d.resourceFactory),
	)

	if !d.strictMode {
		middleware = append(middleware,
			middlewares.NewAwsIamPolicyAttachmentDefaults(),
			middlewares.NewAwsIamRolePolicyDefaults(),
			middlewares.NewAwsIamRoleDefaults(),
		)
	}

	logrus.Debug("Ready to run middlewares")
	err = middleware.Execute(&remoteResources, &resourcesFromState)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to run middlewares")
	}

	if d.filter != nil {
		engine := filter.NewFilterEngine(d.filter)
		remoteResources, err = engine.Run(remoteResources)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to filter remote resources")
		}
		resourcesFromState, err = engine.Run(resourcesFromState)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to filter state resources")
		}
	}

	logrus.Debug("Checking for driftignore")
	driftIgnore := filter.NewDriftIgnore()

	analysis, err := d.analyzer.Analyze(remoteResources, resourcesFromState, driftIgnore)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to perform resources analysis")
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
	resourcesFromState, err = d.iacSupplier.Resources()
	if err != nil {
		return nil, nil, err
	}

	logrus.Info("Start scanning cloud provider")
	remoteResources, err = d.remoteSupplier.Resources()
	if err != nil {
		return nil, nil, err
	}

	return remoteResources, resourcesFromState, err
}
