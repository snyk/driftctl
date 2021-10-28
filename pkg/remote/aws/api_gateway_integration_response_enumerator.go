package aws

import (
	"strings"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type ApiGatewayIntegrationResponseEnumerator struct {
	repository repository.ApiGatewayRepository
	factory    resource.ResourceFactory
}

func NewApiGatewayIntegrationResponseEnumerator(repo repository.ApiGatewayRepository, factory resource.ResourceFactory) *ApiGatewayIntegrationResponseEnumerator {
	return &ApiGatewayIntegrationResponseEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayIntegrationResponseEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayIntegrationResponseResourceType
}

func (e *ApiGatewayIntegrationResponseEnumerator) Enumerate() ([]*resource.Resource, error) {
	apis, err := e.repository.ListAllRestApis()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayRestApiResourceType)
	}

	results := make([]*resource.Resource, 0)

	for _, api := range apis {
		a := api
		resources, err := e.repository.ListAllRestApiResources(*a.Id)
		if err != nil {
			return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayResourceResourceType)
		}

		for _, resource := range resources {
			r := resource
			for httpMethod, method := range r.ResourceMethods {
				if method.MethodIntegration != nil {
					for statusCode := range method.MethodIntegration.IntegrationResponses {
						results = append(
							results,
							e.factory.CreateAbstractResource(
								string(e.SupportedType()),
								strings.Join([]string{"agir", *a.Id, *r.Id, httpMethod, statusCode}, "-"),
								map[string]interface{}{},
							),
						)
					}
				}
			}
		}
	}

	return results, err
}
