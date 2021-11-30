package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type ApiGatewayV2RouteEnumerator struct {
	repository repository.ApiGatewayV2Repository
	factory    resource.ResourceFactory
}

func NewApiGatewayV2RouteEnumerator(repo repository.ApiGatewayV2Repository, factory resource.ResourceFactory) *ApiGatewayV2RouteEnumerator {
	return &ApiGatewayV2RouteEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayV2RouteEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayV2RouteResourceType
}

func (e *ApiGatewayV2RouteEnumerator) Enumerate() ([]*resource.Resource, error) {
	apis, err := e.repository.ListAllApis()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayV2ApiResourceType)
	}

	var results []*resource.Resource
	for _, api := range apis {
		routes, err := e.repository.ListAllApiRoutes(api.ApiId)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}
		for _, route := range routes {
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*route.RouteId,
					map[string]interface{}{
						"api_id":    *api.ApiId,
						"route_key": *route.RouteKey,
					},
				),
			)
		}
	}
	return results, err
}
