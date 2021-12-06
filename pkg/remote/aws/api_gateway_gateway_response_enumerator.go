package aws

import (
	"strings"

	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type ApiGatewayGatewayResponseEnumerator struct {
	repository repository.ApiGatewayRepository
	factory    resource.ResourceFactory
}

func NewApiGatewayGatewayResponseEnumerator(repo repository.ApiGatewayRepository, factory resource.ResourceFactory) *ApiGatewayGatewayResponseEnumerator {
	return &ApiGatewayGatewayResponseEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayGatewayResponseEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayGatewayResponseResourceType
}

func (e *ApiGatewayGatewayResponseEnumerator) Enumerate() ([]*resource.Resource, error) {
	apis, err := e.repository.ListAllRestApis()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayRestApiResourceType)
	}

	results := make([]*resource.Resource, 0)

	for _, api := range apis {
		a := api
		gtwResponses, err := e.repository.ListAllRestApiGatewayResponses(*a.Id)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}

		for _, gtwResponse := range gtwResponses {
			g := gtwResponse
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					strings.Join([]string{"aggr", *a.Id, *g.ResponseType}, "-"),
					map[string]interface{}{},
				),
			)
		}

	}
	return results, err
}
