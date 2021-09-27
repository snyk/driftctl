package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
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

	results := make([]*resource.Resource, len(keys))

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
