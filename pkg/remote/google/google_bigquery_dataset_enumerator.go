package google

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/remote/google/repository"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/google"
)

type GoogleBigqueryDatasetEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleBigqueryDatasetEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleBigqueryDatasetEnumerator {
	return &GoogleBigqueryDatasetEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleBigqueryDatasetEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleBigqueryDatasetResourceType
}

func (e *GoogleBigqueryDatasetEnumerator) Enumerate() ([]*resource.Resource, error) {
	resources, err := e.repository.SearchAllDatasets()

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
					"name": res.DisplayName,
				},
			),
		)
	}

	return results, err
}
