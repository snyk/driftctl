package azurerm

import (
	"github.com/snyk/driftctl/enumeration/remote/azurerm/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/azurerm"
)

type AzurermNetworkSecurityGroupEnumerator struct {
	repository repository.NetworkRepository
	factory    resource.ResourceFactory
}

func NewAzurermNetworkSecurityGroupEnumerator(repo repository.NetworkRepository, factory resource.ResourceFactory) *AzurermNetworkSecurityGroupEnumerator {
	return &AzurermNetworkSecurityGroupEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *AzurermNetworkSecurityGroupEnumerator) SupportedType() resource.ResourceType {
	return azurerm.AzureNetworkSecurityGroupResourceType
}

func (e *AzurermNetworkSecurityGroupEnumerator) Enumerate() ([]*resource.Resource, error) {
	securityGroups, err := e.repository.ListAllSecurityGroups()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), azurerm.AzureNetworkSecurityGroupResourceType)
	}

	results := make([]*resource.Resource, 0, len(securityGroups))

	for _, res := range securityGroups {
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
