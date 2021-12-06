package azurerm

import (
	"github.com/snyk/driftctl/pkg/remote/azurerm/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/azurerm"
)

type AzurermRouteTableEnumerator struct {
	repository repository.NetworkRepository
	factory    resource.ResourceFactory
}

func NewAzurermRouteTableEnumerator(repo repository.NetworkRepository, factory resource.ResourceFactory) *AzurermRouteTableEnumerator {
	return &AzurermRouteTableEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *AzurermRouteTableEnumerator) SupportedType() resource.ResourceType {
	return azurerm.AzureRouteTableResourceType
}

func (e *AzurermRouteTableEnumerator) Enumerate() ([]*resource.Resource, error) {
	resources, err := e.repository.ListAllRouteTables()
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
