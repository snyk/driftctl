package google

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/remote/google/repository"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/google"
)

type GoogleStorageBucketEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleStorageBucketEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleStorageBucketEnumerator {
	return &GoogleStorageBucketEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleStorageBucketEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleStorageBucketResourceType
}

func (e *GoogleStorageBucketEnumerator) Enumerate() ([]*resource.Resource, error) {
	resources, err := e.repository.SearchAllBuckets()

	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, len(resources))

	for _, res := range resources {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				res.DisplayName,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
