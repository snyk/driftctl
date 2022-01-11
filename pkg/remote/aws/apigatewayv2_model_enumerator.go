package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type ApiGatewayV2ModelEnumerator struct {
	repository repository.ApiGatewayV2Repository
	factory    resource.ResourceFactory
}

func NewApiGatewayV2ModelEnumerator(repo repository.ApiGatewayV2Repository, factory resource.ResourceFactory) *ApiGatewayV2ModelEnumerator {
	return &ApiGatewayV2ModelEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayV2ModelEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayV2ModelResourceType
}

func (e *ApiGatewayV2ModelEnumerator) Enumerate() ([]*resource.Resource, error) {
	apis, err := e.repository.ListAllApis()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayV2ApiResourceType)
	}

	var results []*resource.Resource
	for _, api := range apis {
		models, err := e.repository.ListAllApiModels(*api.ApiId)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}
		for _, model := range models {
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*model.ModelId,
					map[string]interface{}{
						"name": *model.Name,
					},
				),
			)
		}
	}
	return results, err
}
