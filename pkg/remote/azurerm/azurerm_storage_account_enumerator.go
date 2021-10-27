package azurerm

import (
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/azurerm"
)

type AzurermStorageAccountEnumerator struct {
	repository repository.StorageRespository
	factory    resource.ResourceFactory
}

func NewAzurermStorageAccountEnumerator(repo repository.StorageRespository, factory resource.ResourceFactory) *AzurermStorageAccountEnumerator {
	return &AzurermStorageAccountEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *AzurermStorageAccountEnumerator) SupportedType() resource.ResourceType {
	return azurerm.AzureStorageAccountResourceType
}

func (e *AzurermStorageAccountEnumerator) Enumerate() ([]*resource.Resource, error) {
	accounts, err := e.repository.ListAllStorageAccount()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(accounts))

	for _, account := range accounts {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*account.ID,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
