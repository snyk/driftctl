package azurerm

import (
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/azurerm"
)

type AzurermSubnetEnumerator struct {
	repository repository.NetworkRepository
	factory    resource.ResourceFactory
}

func NewAzurermSubnetEnumerator(repo repository.NetworkRepository, factory resource.ResourceFactory) *AzurermSubnetEnumerator {
	return &AzurermSubnetEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *AzurermSubnetEnumerator) SupportedType() resource.ResourceType {
	return azurerm.AzureSubnetResourceType
}

func (e *AzurermSubnetEnumerator) Enumerate() ([]*resource.Resource, error) {
	networks, err := e.repository.ListAllVirtualNetworks()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), azurerm.AzureVirtualNetworkResourceType)
	}

	results := make([]*resource.Resource, 0)
	for _, network := range networks {
		resources, err := e.repository.ListAllSubnets(network)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}
		for _, res := range resources {
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*res.ID,
					map[string]interface{}{},
				),
			)
		}
	}

	return results, err
}
