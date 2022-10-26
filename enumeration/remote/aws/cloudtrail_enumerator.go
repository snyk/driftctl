package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type CloudtrailEnumerator struct {
	repository repository.CloudtrailRepository
	factory    resource.ResourceFactory
}

func NewCloudtrailEnumerator(repo repository.CloudtrailRepository, factory resource.ResourceFactory) *CloudtrailEnumerator {
	return &CloudtrailEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *CloudtrailEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsCloudtrailResourceType
}

func (e *CloudtrailEnumerator) Enumerate() ([]*resource.Resource, error) {
	trails, err := e.repository.ListAllTrails()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(trails))

	for _, trail := range trails {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*trail.Name,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
