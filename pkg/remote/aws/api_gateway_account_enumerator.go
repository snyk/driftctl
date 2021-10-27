package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type ApiGatewayAccountEnumerator struct {
	repository repository.ApiGatewayRepository
	factory    resource.ResourceFactory
}

func NewApiGatewayAccountEnumerator(repo repository.ApiGatewayRepository, factory resource.ResourceFactory) *ApiGatewayAccountEnumerator {
	return &ApiGatewayAccountEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayAccountEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayAccountResourceType
}

func (e *ApiGatewayAccountEnumerator) Enumerate() ([]*resource.Resource, error) {
	account, err := e.repository.GetAccount()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, 1)

	if account != nil {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				"api-gateway-account",
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
