package aws

import (
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type CloudfrontDistributionSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       repository.CloudfrontRepository
	runner       *terraform.ParallelResourceReader
}

func NewCloudfrontDistributionSupplier(provider *AWSTerraformProvider) *CloudfrontDistributionSupplier {
	return &CloudfrontDistributionSupplier{
		provider,
		awsdeserializer.NewCloudfrontDistributionDeserializer(),
		repository.NewCloudfrontClient(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s CloudfrontDistributionSupplier) Resources() ([]resource.Resource, error) {
	distributions, err := s.client.ListAllDistributions()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsCloudfrontDistributionResourceType)
	}

	for _, distribution := range distributions {
		d := *distribution
		s.runner.Run(func() (cty.Value, error) {
			return s.readCloudfrontDistribution(d)
		})
	}

	resources, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(resources)
}

func (s CloudfrontDistributionSupplier) readCloudfrontDistribution(distribution cloudfront.DistributionSummary) (cty.Value, error) {
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		ID: *distribution.Id,
		Ty: aws.AwsCloudfrontDistributionResourceType,
	})
	if err != nil {
		logrus.Error(err)
		return cty.NilVal, err
	}
	return *val, nil
}
