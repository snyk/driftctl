package aws

import (
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"

	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type Route53HealthCheckSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       repository.Route53Repository
	runner       *terraform.ParallelResourceReader
}

func NewRoute53HealthCheckSupplier(provider *AWSTerraformProvider) *Route53HealthCheckSupplier {
	return &Route53HealthCheckSupplier{
		provider,
		awsdeserializer.NewRoute53HealthCheckDeserializer(),
		repository.NewRoute53Repository(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s Route53HealthCheckSupplier) Resources() ([]resource.Resource, error) {
	healthChecks, err := s.client.ListAllHealthChecks()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsRoute53HealthCheckResourceType)
	}

	for _, healthCheck := range healthChecks {
		healthCheck := healthCheck
		s.runner.Run(func() (cty.Value, error) {
			return s.readHealthCheck(healthCheck)
		})
	}

	retrieve, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(retrieve)
}

func (s Route53HealthCheckSupplier) readHealthCheck(healthCheck *route53.HealthCheck) (cty.Value, error) {
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		ID: *healthCheck.Id,
		Ty: aws.AwsRoute53HealthCheckResourceType,
	})
	if err != nil {
		logrus.Error(err)
		return cty.NilVal, err
	}
	return *val, nil
}
