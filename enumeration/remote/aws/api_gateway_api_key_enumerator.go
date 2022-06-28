package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type ApiGatewayApiKeyEnumerator struct {
	repository repository.ApiGatewayRepository
	factory    resource.ResourceFactory
}

func NewApiGatewayApiKeyEnumerator(repo repository.ApiGatewayRepository, factory resource.ResourceFactory) *ApiGatewayApiKeyEnumerator {
	return &ApiGatewayApiKeyEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayApiKeyEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayApiKeyResourceType
}

func (e *ApiGatewayApiKeyEnumerator) Enumerate() ([]*resource.Resource, error) {
	keys, err := e.repository.ListAllApiKeys()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(keys))

	for _, key := range keys {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*key.Id,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
