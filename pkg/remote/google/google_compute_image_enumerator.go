package google

import (
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/remote/google/repository"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/google"
)

type GoogleComputeImageEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleComputeImageEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleComputeImageEnumerator {
	return &GoogleComputeImageEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleComputeImageEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleComputeImageResourceType
}

func (e *GoogleComputeImageEnumerator) Enumerate() ([]*resource.Resource, error) {
	resources, err := e.repository.SearchAllImages()

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
				map[string]interface{}{
					"name": res.GetDisplayName(),
				},
			),
		)
	}

	return results, err
}
