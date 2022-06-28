package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"strings"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type ApiGatewayBasePathMappingEnumerator struct {
	repository repository.ApiGatewayRepository
	factory    resource.ResourceFactory
}

func NewApiGatewayBasePathMappingEnumerator(repo repository.ApiGatewayRepository, factory resource.ResourceFactory) *ApiGatewayBasePathMappingEnumerator {
	return &ApiGatewayBasePathMappingEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayBasePathMappingEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayBasePathMappingResourceType
}

func (e *ApiGatewayBasePathMappingEnumerator) Enumerate() ([]*resource.Resource, error) {
	domainNames, err := e.repository.ListAllDomainNames()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayDomainNameResourceType)
	}

	results := make([]*resource.Resource, 0)

	for _, domainName := range domainNames {
		d := domainName
		mappings, err := e.repository.ListAllDomainNameBasePathMappings(*d.DomainName)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}

		for _, mapping := range mappings {
			m := mapping

			basePath := ""
			if m.BasePath != nil && *m.BasePath != "(none)" {
				basePath = *m.BasePath
			}

			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					strings.Join([]string{*d.DomainName, basePath}, "/"),
					map[string]interface{}{},
				),
			)
		}

	}

	return results, err
}
