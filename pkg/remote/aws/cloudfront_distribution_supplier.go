package aws

import (
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type CloudfrontDistributionSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	client       repository.CloudfrontRepository
	runner       *terraform.ParallelResourceReader
}

func NewCloudfrontDistributionSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.CloudfrontRepository) *CloudfrontDistributionSupplier {
	return &CloudfrontDistributionSupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *CloudfrontDistributionSupplier) SuppliedType() resource.ResourceType {
	return aws.AwsCloudfrontDistributionResourceType
}

func (s *CloudfrontDistributionSupplier) Resources() ([]resource.Resource, error) {
	distributions, err := s.client.ListAllDistributions()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
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

	return s.deserializer.Deserialize(s.SuppliedType(), resources)
}

func (s *CloudfrontDistributionSupplier) readCloudfrontDistribution(distribution cloudfront.DistributionSummary) (cty.Value, error) {
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		ID: *distribution.Id,
		Ty: s.SuppliedType(),
	})
	if err != nil {
		logrus.Error(err)
		return cty.NilVal, err
	}
	return *val, nil
}
