package azurerm

import (
	"github.com/snyk/driftctl/enumeration/remote/azurerm/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/azurerm"
)

type AzurermPrivateDNSZoneEnumerator struct {
	repository repository.PrivateDNSRepository
	factory    resource.ResourceFactory
}

func NewAzurermPrivateDNSZoneEnumerator(repo repository.PrivateDNSRepository, factory resource.ResourceFactory) *AzurermPrivateDNSZoneEnumerator {
	return &AzurermPrivateDNSZoneEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *AzurermPrivateDNSZoneEnumerator) SupportedType() resource.ResourceType {
	return azurerm.AzurePrivateDNSZoneResourceType
}

func (e *AzurermPrivateDNSZoneEnumerator) Enumerate() ([]*resource.Resource, error) {

	zones, err := e.repository.ListAllPrivateZones()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0)

	for _, zone := range zones {

		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*zone.ID,
				map[string]interface{}{},
			),
		)

	}

	return results, err
}
