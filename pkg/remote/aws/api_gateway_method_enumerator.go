package aws

import (
	"strings"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type ApiGatewayMethodEnumerator struct {
	repository repository.ApiGatewayRepository
	factory    resource.ResourceFactory
}

func NewApiGatewayMethodEnumerator(repo repository.ApiGatewayRepository, factory resource.ResourceFactory) *ApiGatewayMethodEnumerator {
	return &ApiGatewayMethodEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayMethodEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayMethodResourceType
}

func (e *ApiGatewayMethodEnumerator) Enumerate() ([]*resource.Resource, error) {
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
			for method := range r.ResourceMethods {
				results = append(
					results,
					e.factory.CreateAbstractResource(
						string(e.SupportedType()),
						strings.Join([]string{"agm", *a.Id, *r.Id, method}, "-"),
						map[string]interface{}{},
					),
				)
			}
		}
	}

	return results, err
}
