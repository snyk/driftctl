package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type ApiGatewayAuthorizerEnumerator struct {
	repository repository.ApiGatewayRepository
	factory    resource.ResourceFactory
}

func NewApiGatewayAuthorizerEnumerator(repo repository.ApiGatewayRepository, factory resource.ResourceFactory) *ApiGatewayAuthorizerEnumerator {
	return &ApiGatewayAuthorizerEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayAuthorizerEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayAuthorizerResourceType
}

func (e *ApiGatewayAuthorizerEnumerator) Enumerate() ([]*resource.Resource, error) {
	apis, err := e.repository.ListAllRestApis()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayRestApiResourceType)
	}

	results := make([]*resource.Resource, 0)

	for _, api := range apis {
		a := api
		authorizers, err := e.repository.ListAllRestApiAuthorizers(*a.Id)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}

		for _, authorizer := range authorizers {
			au := authorizer
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*au.Id,
					map[string]interface{}{},
				),
			)
		}

	}

	return results, err
}
