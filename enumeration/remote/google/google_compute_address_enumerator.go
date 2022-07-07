package google

import (
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/remote/google/repository"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/google"
)

type GoogleComputeAddressEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleComputeAddressEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleComputeAddressEnumerator {
	return &GoogleComputeAddressEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleComputeAddressEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleComputeAddressResourceType
}

func (e *GoogleComputeAddressEnumerator) Enumerate() ([]*resource.Resource, error) {
	resources, err := e.repository.SearchAllAddresses()

	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(resources))

	for _, res := range resources {
		// Global addresses are handled as a dedicated resource
		if res.GetLocation() == "global" {
			continue
		}
		address := ""
		if addr, exist := res.GetAdditionalAttributes().GetFields()["address"]; exist {
			address = addr.GetStringValue()
		}
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				trimResourceName(res.GetName()),
				map[string]interface{}{
					"name":    res.GetDisplayName(),
					"address": address,
				},
			),
		)
	}

	return results, err
}
