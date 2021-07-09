package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
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

func (e *LambdaEventSourceMappingEnumerator) Enumerate() ([]resource.Resource, error) {
	eventSourceMappings, err := e.repository.ListAllLambdaEventSourceMappings()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, string(e.SupportedType()))
	}

	results := make([]resource.Resource, len(eventSourceMappings))

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
