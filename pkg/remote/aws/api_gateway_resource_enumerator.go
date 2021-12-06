package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type ApiGatewayResourceEnumerator struct {
	repository repository.ApiGatewayRepository
	factory    resource.ResourceFactory
}

func NewApiGatewayResourceEnumerator(repo repository.ApiGatewayRepository, factory resource.ResourceFactory) *ApiGatewayResourceEnumerator {
	return &ApiGatewayResourceEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayResourceEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayResourceResourceType
}

func (e *ApiGatewayResourceEnumerator) Enumerate() ([]*resource.Resource, error) {
	apis, err := e.repository.ListAllRestApis()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayRestApiResourceType)
	}

	results := make([]*resource.Resource, 0)

	for _, api := range apis {
		a := api
		resources, err := e.repository.ListAllRestApiResources(*a.Id)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}

		for _, resource := range resources {
			r := resource
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*r.Id,
					map[string]interface{}{
						"rest_api_id": *a.Id,
						"path":        *r.Path,
					},
				),
			)
		}
	}

	return results, err
}
