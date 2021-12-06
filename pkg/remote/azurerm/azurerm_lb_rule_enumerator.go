package azurerm

import (
	"github.com/snyk/driftctl/pkg/remote/azurerm/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/azurerm"
)

type AzurermLoadBalancerRuleEnumerator struct {
	repository repository.NetworkRepository
	factory    resource.ResourceFactory
}

func NewAzurermLoadBalancerRuleEnumerator(repo repository.NetworkRepository, factory resource.ResourceFactory) *AzurermLoadBalancerRuleEnumerator {
	return &AzurermLoadBalancerRuleEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *AzurermLoadBalancerRuleEnumerator) SupportedType() resource.ResourceType {
	return azurerm.AzureLoadBalancerRuleResourceType
}

func (e *AzurermLoadBalancerRuleEnumerator) Enumerate() ([]*resource.Resource, error) {
	loadBalancers, err := e.repository.ListAllLoadBalancers()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), azurerm.AzureLoadBalancerResourceType)
	}

	results := make([]*resource.Resource, 0)

	for _, res := range loadBalancers {
		rules, err := e.repository.ListLoadBalancerRules(res)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}

		for _, rule := range rules {
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*rule.ID,
					map[string]interface{}{
						"name":            *rule.Name,
						"loadbalancer_id": *res.ID,
					},
				),
			)
		}
	}

	return results, err
}
