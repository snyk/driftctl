package google

import (
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/remote/google/repository"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/google"
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
					"name":     res.GetDisplayName(),
					"location": res.GetLocation(),
				},
			),
		)
	}

	return results, err
}
