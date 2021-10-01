package google

import (
	"fmt"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/remote/google/repository"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/google"
)

type GoogleStorageBucketIamBindingEnumerator struct {
	repository        repository.AssetRepository
	storageRepository repository.StorageRepository
	factory           resource.ResourceFactory
}

func NewGoogleStorageBucketIamBindingEnumerator(repo repository.AssetRepository, storageRepo repository.StorageRepository, factory resource.ResourceFactory) *GoogleStorageBucketIamBindingEnumerator {
	return &GoogleStorageBucketIamBindingEnumerator{
		repository:        repo,
		storageRepository: storageRepo,
		factory:           factory,
	}
}

func (e *GoogleStorageBucketIamBindingEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleStorageBucketIamBindingResourceType
}

func (e *GoogleStorageBucketIamBindingEnumerator) Enumerate() ([]*resource.Resource, error) {
	resources, err := e.repository.SearchAllBuckets()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), google.GoogleStorageBucketResourceType)
	}

	results := make([]*resource.Resource, len(resources))

	for _, bucket := range resources {
		bindings, err := e.storageRepository.ListAllBindings(bucket.DisplayName)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}
		for roleName, members := range bindings {
			id := fmt.Sprintf("b/%s/%s", bucket.DisplayName, roleName)
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					id,
					map[string]interface{}{
						"id":      id,
						"bucket":  fmt.Sprintf("b/%s", bucket.DisplayName),
						"role":    roleName,
						"members": members,
					},
				),
			)
		}
	}

	return results, err
}
