package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type CloudfrontDistributionEnumerator struct {
	repository repository.CloudfrontRepository
	factory    resource.ResourceFactory
}

func NewCloudfrontDistributionEnumerator(repo repository.CloudfrontRepository, factory resource.ResourceFactory) *CloudfrontDistributionEnumerator {
	return &CloudfrontDistributionEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *CloudfrontDistributionEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsCloudfrontDistributionResourceType
}

func (e *CloudfrontDistributionEnumerator) Enumerate() ([]resource.Resource, error) {
	distributions, err := e.repository.ListAllDistributions()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]resource.Resource, len(distributions))

	for _, distribution := range distributions {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*distribution.Id,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
