package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	resourceaws "github.com/snyk/driftctl/enumeration/resource/aws"
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

	results := make([]*resource.Resource, 0, len(functions))

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
