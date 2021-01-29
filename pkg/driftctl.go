package pkg

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/middlewares"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/jmespath/go-jmespath"
	"github.com/sirupsen/logrus"
)

type DriftCTL struct {
	remoteSupplier resource.Supplier
	iacSupplier    resource.Supplier
	analyzer       analyser.Analyzer
	filter         *jmespath.JMESPath
}

func NewDriftCTL(remoteSupplier resource.Supplier, iacSupplier resource.Supplier, filter *jmespath.JMESPath, alerter *alerter.Alerter) *DriftCTL {
	return &DriftCTL{remoteSupplier, iacSupplier, analyser.NewAnalyzer(alerter), filter}
}

func (d DriftCTL) Run() *analyser.Analysis {
	remoteResources, resourcesFromState, err := d.scan()
	if err != nil {
		logrus.Errorf("Unable to scan resources: %+v", err)
		return nil
	}

	middleware := middlewares.NewChain(
		middlewares.NewRoute53DefaultZoneRecordSanitizer(),
		middlewares.NewS3BucketAcl(),
		middlewares.NewAwsInstanceBlockDeviceResourceMapper(),
		middlewares.NewVPCDefaultSecurityGroupSanitizer(),
		middlewares.NewVPCSecurityGroupRuleSanitizer(),
		middlewares.NewIamPolicyAttachmentSanitizer(),
		middlewares.AwsInstanceEIP{},
		middlewares.NewAwsDefaultInternetGatewayRoute(),
		middlewares.NewAwsDefaultInternetGateway(),
		middlewares.NewAwsDefaultVPC(),
		middlewares.NewAwsDefaultSubnet(),
		middlewares.NewAwsRouteTableExpander(),
		middlewares.NewAwsDefaultRouteTable(),
		middlewares.NewAwsDefaultRoute(),
		middlewares.NewAwsNatGatewayEipAssoc(),
		middlewares.NewAwsBucketPolicyExpander(),
		middlewares.NewAwsSqsQueuePolicyExpander(),
		middlewares.NewAwsDefaultSqsQueuePolicy(),
	)

	logrus.Debug("Ready to run middlewares")
	err = middleware.Execute(&remoteResources, &resourcesFromState)
	if err != nil {
		logrus.Errorf("Unable to run middlewares: %+v", err)
		return nil
	}

	if d.filter != nil {
		engine := filter.NewFilterEngine(d.filter)
		remoteResources, err = engine.Run(remoteResources)
		if err != nil {
			logrus.Error(err)
			return nil
		}
		resourcesFromState, err = engine.Run(resourcesFromState)
		if err != nil {
			logrus.Error(err)
			return nil
		}
	}

	logrus.Debug("Checking for driftignore")
	driftIgnore := filter.NewDriftIgnore()

	analysis, err := d.analyzer.Analyze(remoteResources, resourcesFromState, driftIgnore)

	if err != nil {
		logrus.Errorf("Unable to analyse resources: %+v", err)
		return nil
	}

	return &analysis
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
