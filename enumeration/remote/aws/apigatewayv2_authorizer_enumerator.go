package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type ApiGatewayV2AuthorizerEnumerator struct {
	repository repository.ApiGatewayV2Repository
	factory    resource.ResourceFactory
}

func NewApiGatewayV2AuthorizerEnumerator(repo repository.ApiGatewayV2Repository, factory resource.ResourceFactory) *ApiGatewayV2AuthorizerEnumerator {
	return &ApiGatewayV2AuthorizerEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayV2AuthorizerEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayV2AuthorizerResourceType
}

func (e *ApiGatewayV2AuthorizerEnumerator) Enumerate() ([]*resource.Resource, error) {
	apis, err := e.repository.ListAllApis()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayV2ApiResourceType)
	}

	results := make([]*resource.Resource, 0)

	for _, api := range apis {
		a := api
		authorizers, err := e.repository.ListAllApiAuthorizers(*a.ApiId)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}

		for _, authorizer := range authorizers {
			au := authorizer
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*au.AuthorizerId,
					map[string]interface{}{},
				),
			)
		}

	}

	return results, err
}
