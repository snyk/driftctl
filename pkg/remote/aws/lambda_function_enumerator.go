package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type LambdaFunctionEnumerator struct {
	repository repository.LambdaRepository
	factory    resource.ResourceFactory
}

func NewLambdaFunctionEnumerator(repo repository.LambdaRepository, factory resource.ResourceFactory) *LambdaFunctionEnumerator {
	return &LambdaFunctionEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *LambdaFunctionEnumerator) SupportedType() resource.ResourceType {
	return resourceaws.AwsLambdaFunctionResourceType
}

func (e *LambdaFunctionEnumerator) Enumerate() ([]*resource.Resource, error) {
	functions, err := e.repository.ListAllLambdaFunctions()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, len(functions))

	for _, function := range functions {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*function.FunctionName,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
