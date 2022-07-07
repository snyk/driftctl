package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	resourceaws "github.com/snyk/driftctl/enumeration/resource/aws"
)

type LambdaEventSourceMappingEnumerator struct {
	repository repository.LambdaRepository
	factory    resource.ResourceFactory
}

func NewLambdaEventSourceMappingEnumerator(repo repository.LambdaRepository, factory resource.ResourceFactory) *LambdaEventSourceMappingEnumerator {
	return &LambdaEventSourceMappingEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *LambdaEventSourceMappingEnumerator) SupportedType() resource.ResourceType {
	return resourceaws.AwsLambdaEventSourceMappingResourceType
}

func (e *LambdaEventSourceMappingEnumerator) Enumerate() ([]*resource.Resource, error) {
	eventSourceMappings, err := e.repository.ListAllLambdaEventSourceMappings()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(eventSourceMappings))

	for _, eventSourceMapping := range eventSourceMappings {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*eventSourceMapping.UUID,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
