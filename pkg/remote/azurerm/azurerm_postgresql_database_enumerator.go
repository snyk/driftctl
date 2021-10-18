package azurerm

import (
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/azurerm"
	"github.com/sirupsen/logrus"
)

type AzurermPostgresqlDatabaseEnumerator struct {
	repository repository.PostgresqlRespository
	factory    resource.ResourceFactory
}

func NewAzurermPostgresqlDatabaseEnumerator(repo repository.PostgresqlRespository, factory resource.ResourceFactory) *AzurermPostgresqlDatabaseEnumerator {
	return &AzurermPostgresqlDatabaseEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *AzurermPostgresqlDatabaseEnumerator) SupportedType() resource.ResourceType {
	return azurerm.AzurePostgresqlDatabaseResourceType
}

func (e *AzurermPostgresqlDatabaseEnumerator) Enumerate() ([]*resource.Resource, error) {
	servers, err := e.repository.ListAllServers()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), azurerm.AzurePostgresqlServerResourceType)
	}

	results := make([]*resource.Resource, 0)
	for _, server := range servers {
		res, err := azure.ParseResourceID(*server.ID)
		if err != nil {
			logrus.WithFields(map[string]interface{}{
				"type": azurerm.AzurePostgresqlServerResourceType,
				"id":   *server.ID,
			}).Errorf("Error listing %s: failed to parse resource ID", string(e.SupportedType()))
			continue
		}

		databases, err := e.repository.ListAllDatabasesByServer(res.ResourceGroup, server)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}

		for _, db := range databases {
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*db.ID,
					map[string]interface{}{
						"name": *db.Name,
					},
				),
			)
		}
	}

	return results, err
}
