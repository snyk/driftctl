package google

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/remote/google/repository"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/google"
)

type GoogleComputeRouteEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleComputeRouteEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleComputeRouteEnumerator {
	return &GoogleComputeRouteEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleComputeRouteEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleComputeRouteResourceType
}

func (e *GoogleComputeRouteEnumerator) Enumerate() ([]*resource.Resource, error) {
	resources, err := e.repository.SearchAllRoutes()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(resources))

	for _, res := range resources {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				trimResourceName(res.GetName()),
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
