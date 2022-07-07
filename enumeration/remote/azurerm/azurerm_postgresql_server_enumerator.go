package azurerm

import (
	"github.com/snyk/driftctl/enumeration/remote/azurerm/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/azurerm"
)

type AzurermPostgresqlServerEnumerator struct {
	repository repository.PostgresqlRespository
	factory    resource.ResourceFactory
}

func NewAzurermPostgresqlServerEnumerator(repo repository.PostgresqlRespository, factory resource.ResourceFactory) *AzurermPostgresqlServerEnumerator {
	return &AzurermPostgresqlServerEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *AzurermPostgresqlServerEnumerator) SupportedType() resource.ResourceType {
	return azurerm.AzurePostgresqlServerResourceType
}

func (e *AzurermPostgresqlServerEnumerator) Enumerate() ([]*resource.Resource, error) {
	servers, err := e.repository.ListAllServers()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0)
	for _, server := range servers {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*server.ID,
				map[string]interface{}{
					"name": *server.Name,
				},
			),
		)
	}

	return results, err
}
