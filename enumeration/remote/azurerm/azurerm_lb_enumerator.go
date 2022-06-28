package azurerm

import (
	"github.com/snyk/driftctl/enumeration/remote/azurerm/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/azurerm"
)

type AzurermLoadBalancerEnumerator struct {
	repository repository.NetworkRepository
	factory    resource.ResourceFactory
}

func NewAzurermLoadBalancerEnumerator(repo repository.NetworkRepository, factory resource.ResourceFactory) *AzurermLoadBalancerEnumerator {
	return &AzurermLoadBalancerEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *AzurermLoadBalancerEnumerator) SupportedType() resource.ResourceType {
	return azurerm.AzureLoadBalancerResourceType
}

func (e *AzurermLoadBalancerEnumerator) Enumerate() ([]*resource.Resource, error) {
	loadBalancers, err := e.repository.ListAllLoadBalancers()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(loadBalancers))

	for _, res := range loadBalancers {
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
