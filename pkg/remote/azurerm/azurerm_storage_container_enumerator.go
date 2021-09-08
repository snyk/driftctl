package azurerm

import (
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/azurerm"
)

type AzurermStorageContainerEnumerator struct {
	repository repository.StorageRespository
	factory    resource.ResourceFactory
}

func NewAzurermStorageContainerEnumerator(repo repository.StorageRespository, factory resource.ResourceFactory) *AzurermStorageContainerEnumerator {
	return &AzurermStorageContainerEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *AzurermStorageContainerEnumerator) SupportedType() resource.ResourceType {
	return azurerm.AzureStorageContainerResourceType
}

func (e *AzurermStorageContainerEnumerator) Enumerate() ([]*resource.Resource, error) {

	accounts, err := e.repository.ListAllStorageAccount()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), azurerm.AzureStorageAccountResourceType)
	}

	results := make([]*resource.Resource, 0)

	for _, account := range accounts {
		containers, err := e.repository.ListAllStorageContainer(account)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}

		for _, container := range containers {
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					container,
					map[string]interface{}{},
				),
			)
		}
	}

	return results, err
}
