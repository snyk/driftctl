package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type ApiGatewayModelEnumerator struct {
	repository repository.ApiGatewayRepository
	factory    resource.ResourceFactory
}

func NewApiGatewayModelEnumerator(repo repository.ApiGatewayRepository, factory resource.ResourceFactory) *ApiGatewayModelEnumerator {
	return &ApiGatewayModelEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayModelEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayModelResourceType
}

func (e *ApiGatewayModelEnumerator) Enumerate() ([]*resource.Resource, error) {
	apis, err := e.repository.ListAllRestApis()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayRestApiResourceType)
	}

	results := make([]*resource.Resource, 0)

	for _, api := range apis {
		a := api
		models, err := e.repository.ListAllRestApiModels(*a.Id)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}

		for _, model := range models {
			m := model
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*m.Id,
					map[string]interface{}{},
				),
			)
		}
	}

	return results, err
}
