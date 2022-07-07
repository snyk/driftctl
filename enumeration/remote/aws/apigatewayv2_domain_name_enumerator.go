package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type ApiGatewayV2DomainNameEnumerator struct {
	// AWS SDK list domain names endpoint from API Gateway v2 returns the
	// same results as the v1 one, thus let's re-use the method from
	// the API Gateway v1
	repository repository.ApiGatewayRepository
	factory    resource.ResourceFactory
}

func NewApiGatewayV2DomainNameEnumerator(repo repository.ApiGatewayRepository, factory resource.ResourceFactory) *ApiGatewayV2DomainNameEnumerator {
	return &ApiGatewayV2DomainNameEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayV2DomainNameEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayV2DomainNameResourceType
}

func (e *ApiGatewayV2DomainNameEnumerator) Enumerate() ([]*resource.Resource, error) {
	domainNames, err := e.repository.ListAllDomainNames()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(domainNames))

	for _, domainName := range domainNames {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*domainName.DomainName,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
