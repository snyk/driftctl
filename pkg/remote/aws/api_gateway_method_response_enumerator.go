package aws

import (
	"strings"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type ApiGatewayMethodResponseEnumerator struct {
	repository repository.ApiGatewayRepository
	factory    resource.ResourceFactory
}

func NewApiGatewayMethodResponseEnumerator(repo repository.ApiGatewayRepository, factory resource.ResourceFactory) *ApiGatewayMethodResponseEnumerator {
	return &ApiGatewayMethodResponseEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayMethodResponseEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayMethodResponseResourceType
}

func (e *ApiGatewayMethodResponseEnumerator) Enumerate() ([]*resource.Resource, error) {
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
				for statusCode := range method.MethodResponses {
					results = append(
						results,
						e.factory.CreateAbstractResource(
							string(e.SupportedType()),
							strings.Join([]string{"agmr", *a.Id, *r.Id, httpMethod, statusCode}, "-"),
							map[string]interface{}{},
						),
					)
				}
			}
		}
	}

	return results, err
}
