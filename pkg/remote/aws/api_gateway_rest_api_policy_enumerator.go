package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type ApiGatewayRestApiPolicyEnumerator struct {
	repository repository.ApiGatewayRepository
	factory    resource.ResourceFactory
}

func NewApiGatewayRestApiPolicyEnumerator(repo repository.ApiGatewayRepository, factory resource.ResourceFactory) *ApiGatewayRestApiPolicyEnumerator {
	return &ApiGatewayRestApiPolicyEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayRestApiPolicyEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayRestApiPolicyResourceType
}

func (e *ApiGatewayRestApiPolicyEnumerator) Enumerate() ([]*resource.Resource, error) {
	apis, err := e.repository.ListAllRestApis()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayRestApiResourceType)
	}

	results := make([]*resource.Resource, 0)

	for _, api := range apis {
		a := api
		if a.Policy == nil || *a.Policy == "" {
			continue
		}
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*a.Id,
				map[string]interface{}{},
			),
		)
	}
	return results, err
}
