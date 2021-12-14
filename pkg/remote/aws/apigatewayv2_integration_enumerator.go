package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type ApiGatewayV2IntegrationEnumerator struct {
	repository repository.ApiGatewayV2Repository
	factory    resource.ResourceFactory
}

func NewApiGatewayV2IntegrationEnumerator(repo repository.ApiGatewayV2Repository, factory resource.ResourceFactory) *ApiGatewayV2IntegrationEnumerator {
	return &ApiGatewayV2IntegrationEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayV2IntegrationEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayV2IntegrationResourceType
}

func (e *ApiGatewayV2IntegrationEnumerator) Enumerate() ([]*resource.Resource, error) {
	apis, err := e.repository.ListAllApis()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayV2ApiResourceType)
	}

	results := make([]*resource.Resource, 0)

	for _, a := range apis {
		api := a
		integrations, err := e.repository.ListAllApiIntegrations(*api.ApiId)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}

		for _, integration := range integrations {
			data := map[string]interface{}{
				"api_id":           *api.ApiId,
				"integration_type": *integration.IntegrationType,
			}

			if integration.IntegrationMethod != nil {
				// this is needed to discriminate in middleware. But it is nil when the type is mock...
				data["integration_method"] = *integration.IntegrationMethod
			}

			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*integration.IntegrationId,
					data,
				),
			)
		}
	}
	return results, err
}
