package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type ApiGatewayV2MappingEnumerator struct {
	repository   repository.ApiGatewayV2Repository
	repositoryV1 repository.ApiGatewayRepository
	factory      resource.ResourceFactory
}

func NewApiGatewayV2MappingEnumerator(repo repository.ApiGatewayV2Repository, repov1 repository.ApiGatewayRepository, factory resource.ResourceFactory) *ApiGatewayV2MappingEnumerator {
	return &ApiGatewayV2MappingEnumerator{
		repository:   repo,
		repositoryV1: repov1,
		factory:      factory,
	}
}

func (e *ApiGatewayV2MappingEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayV2MappingResourceType
}

func (e *ApiGatewayV2MappingEnumerator) Enumerate() ([]*resource.Resource, error) {
	domainNames, err := e.repositoryV1.ListAllDomainNames()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayDomainNameResourceType)
	}

	var results []*resource.Resource
	for _, domainName := range domainNames {
		mappings, err := e.repository.ListAllApiMappings(*domainName.DomainName)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}
		for _, mapping := range mappings {
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*mapping.ApiMappingId,
					map[string]interface{}{
						"stage":  *mapping.Stage,
						"api_id": *mapping.ApiId,
					},
				),
			)
		}
	}
	return results, err
}
