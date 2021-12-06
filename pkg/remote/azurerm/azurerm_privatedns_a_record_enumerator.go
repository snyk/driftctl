package azurerm

import (
	"github.com/snyk/driftctl/pkg/remote/azurerm/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/azurerm"
)

type AzurermPrivateDNSARecordEnumerator struct {
	repository repository.PrivateDNSRepository
	factory    resource.ResourceFactory
}

func NewAzurermPrivateDNSARecordEnumerator(repo repository.PrivateDNSRepository, factory resource.ResourceFactory) *AzurermPrivateDNSARecordEnumerator {
	return &AzurermPrivateDNSARecordEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *AzurermPrivateDNSARecordEnumerator) SupportedType() resource.ResourceType {
	return azurerm.AzurePrivateDNSARecordResourceType
}

func (e *AzurermPrivateDNSARecordEnumerator) Enumerate() ([]*resource.Resource, error) {

	zones, err := e.repository.ListAllPrivateZones()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), azurerm.AzurePrivateDNSZoneResourceType)
	}

	results := make([]*resource.Resource, 0)

	for _, zone := range zones {
		records, err := e.repository.ListAllARecords(zone)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}
		for _, record := range records {
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*record.ID,
					map[string]interface{}{
						"name":      *record.Name,
						"zone_name": *zone.Name,
					},
				),
			)
		}

	}

	return results, err
}
