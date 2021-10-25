package google

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/remote/google/repository"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/google"
)

type GoogleComputeSubnetworkEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleComputeSubnetworkEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleComputeSubnetworkEnumerator {
	return &GoogleComputeSubnetworkEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleComputeSubnetworkEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleComputeSubnetworkResourceType
}

func (e *GoogleComputeSubnetworkEnumerator) Enumerate() ([]*resource.Resource, error) {
	subnets, err := e.repository.SearchAllSubnetworks()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(subnets))

	for _, res := range subnets {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				trimResourceName(res.GetName()),
				map[string]interface{}{
					"display_name": res.GetDisplayName(),
					"location":     res.GetLocation(),
				},
			),
		)
	}

	return results, err
}
