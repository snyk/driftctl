package azurerm

import (
	"github.com/snyk/driftctl/pkg/remote/azurerm/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/azurerm"
)

type AzurermFirewallsEnumerator struct {
	repository repository.NetworkRepository
	factory    resource.ResourceFactory
}

func NewAzurermFirewallsEnumerator(repo repository.NetworkRepository, factory resource.ResourceFactory) *AzurermFirewallsEnumerator {
	return &AzurermFirewallsEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *AzurermFirewallsEnumerator) SupportedType() resource.ResourceType {
	return azurerm.AzureFirewallResourceType
}

func (e *AzurermFirewallsEnumerator) Enumerate() ([]*resource.Resource, error) {
	resources, err := e.repository.ListAllFirewalls()
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
