package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type ApiGatewayV2IntegrationResponseEnumerator struct {
	repository repository.ApiGatewayV2Repository
	factory    resource.ResourceFactory
}

func NewApiGatewayV2IntegrationResponseEnumerator(repo repository.ApiGatewayV2Repository, factory resource.ResourceFactory) *ApiGatewayV2IntegrationResponseEnumerator {
	return &ApiGatewayV2IntegrationResponseEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayV2IntegrationResponseEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayV2IntegrationResponseResourceType
}

func (e *ApiGatewayV2IntegrationResponseEnumerator) Enumerate() ([]*resource.Resource, error) {
	apis, err := e.repository.ListAllApis()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayV2ApiResourceType)
	}

	results := make([]*resource.Resource, 0)

	for _, a := range apis {
		apiID := *a.ApiId
		integrations, err := e.repository.ListAllApiIntegrations(apiID)
		if err != nil {
			return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayV2IntegrationResourceType)
		}

		for _, integration := range integrations {
			integrationId := *integration.IntegrationId
			responses, err := e.repository.ListAllApiIntegrationResponses(apiID, integrationId)
			if err != nil {
				return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
			}

			for _, resp := range responses {
				responseId := *resp.IntegrationResponseId
				results = append(
					results,
					e.factory.CreateAbstractResource(
						string(e.SupportedType()),
						responseId,
						map[string]interface{}{},
					),
				)
			}

		}
	}
	return results, err
}
