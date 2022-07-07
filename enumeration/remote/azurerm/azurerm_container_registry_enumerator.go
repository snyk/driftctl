package azurerm

import (
	"github.com/snyk/driftctl/enumeration/remote/azurerm/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/azurerm"
)

type AzurermContainerRegistryEnumerator struct {
	repository repository.ContainerRegistryRepository
	factory    resource.ResourceFactory
}

func NewAzurermContainerRegistryEnumerator(repo repository.ContainerRegistryRepository, factory resource.ResourceFactory) *AzurermContainerRegistryEnumerator {
	return &AzurermContainerRegistryEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *AzurermContainerRegistryEnumerator) SupportedType() resource.ResourceType {
	return azurerm.AzureContainerRegistryResourceType
}

func (e *AzurermContainerRegistryEnumerator) Enumerate() ([]*resource.Resource, error) {
	registries, err := e.repository.ListAllContainerRegistries()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0)
	for _, registry := range registries {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*registry.ID,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
