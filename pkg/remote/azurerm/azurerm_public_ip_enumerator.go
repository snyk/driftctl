package azurerm

import (
	"github.com/snyk/driftctl/pkg/remote/azurerm/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/azurerm"
)

type AzurermPublicIPEnumerator struct {
	repository repository.NetworkRepository
	factory    resource.ResourceFactory
}

func NewAzurermPublicIPEnumerator(repo repository.NetworkRepository, factory resource.ResourceFactory) *AzurermPublicIPEnumerator {
	return &AzurermPublicIPEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *AzurermPublicIPEnumerator) SupportedType() resource.ResourceType {
	return azurerm.AzurePublicIPResourceType
}

func (e *AzurermPublicIPEnumerator) Enumerate() ([]*resource.Resource, error) {
	resources, err := e.repository.ListAllPublicIPAddresses()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(resources))

	for _, res := range resources {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*res.ID,
				map[string]interface{}{
					"name": *res.Name,
				},
			),
		)
	}

	return results, err
}
