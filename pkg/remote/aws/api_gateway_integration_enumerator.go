package aws

import (
	"strings"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type ApiGatewayIntegrationEnumerator struct {
	repository repository.ApiGatewayRepository
	factory    resource.ResourceFactory
}

func NewApiGatewayIntegrationEnumerator(repo repository.ApiGatewayRepository, factory resource.ResourceFactory) *ApiGatewayIntegrationEnumerator {
	return &ApiGatewayIntegrationEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayIntegrationEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayIntegrationResourceType
}

func (e *ApiGatewayIntegrationEnumerator) Enumerate() ([]*resource.Resource, error) {
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
			for httpMethod := range r.ResourceMethods {
				results = append(
					results,
					e.factory.CreateAbstractResource(
						string(e.SupportedType()),
						strings.Join([]string{"agi", *a.Id, *r.Id, httpMethod}, "-"),
						map[string]interface{}{},
					),
				)
			}
		}
	}

	return results, err
}
