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

	authorizers, err := e.repository.ListAllRestApiAuthorizers(apis)
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, len(authorizers))

	for _, authorizer := range authorizers {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*authorizer.Id,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
