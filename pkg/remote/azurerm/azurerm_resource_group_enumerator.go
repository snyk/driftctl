package azurerm

import (
	"github.com/snyk/driftctl/pkg/remote/azurerm/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/azurerm"
)

type AzurermResourceGroupEnumerator struct {
	repository repository.ResourcesRepository
	factory    resource.ResourceFactory
}

func NewAzurermResourceGroupEnumerator(repo repository.ResourcesRepository, factory resource.ResourceFactory) *AzurermResourceGroupEnumerator {
	return &AzurermResourceGroupEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *AzurermResourceGroupEnumerator) SupportedType() resource.ResourceType {
	return azurerm.AzureResourceGroupResourceType
}

func (e *AzurermResourceGroupEnumerator) Enumerate() ([]*resource.Resource, error) {
	groups, err := e.repository.ListAllResourceGroups()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0)
	for _, group := range groups {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*group.ID,
				map[string]interface{}{
					"name": *group.Name,
				},
			),
		)
	}

	return results, err
}
