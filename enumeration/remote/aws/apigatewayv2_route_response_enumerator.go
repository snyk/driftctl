package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type ApiGatewayV2RouteResponseEnumerator struct {
	repository repository.ApiGatewayV2Repository
	factory    resource.ResourceFactory
}

func NewApiGatewayV2RouteResponseEnumerator(repo repository.ApiGatewayV2Repository, factory resource.ResourceFactory) *ApiGatewayV2RouteResponseEnumerator {
	return &ApiGatewayV2RouteResponseEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayV2RouteResponseEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayV2RouteResponseResourceType
}

func (e *ApiGatewayV2RouteResponseEnumerator) Enumerate() ([]*resource.Resource, error) {
	apis, err := e.repository.ListAllApis()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayV2ApiResourceType)
	}

	var results []*resource.Resource
	for _, api := range apis {
		a := api
		routes, err := e.repository.ListAllApiRoutes(a.ApiId)
		if err != nil {
			return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayV2RouteResourceType)
		}
		for _, route := range routes {
			r := route
			responses, err := e.repository.ListAllApiRouteResponses(*a.ApiId, *r.RouteId)
			if err != nil {
				return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
			}
			for _, response := range responses {
				res := response
				results = append(
					results,
					e.factory.CreateAbstractResource(
						string(e.SupportedType()),
						*res.RouteResponseId,
						map[string]interface{}{},
					),
				)
			}
		}
	}
	return results, err
}
