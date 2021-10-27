package google

import (
	"fmt"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/remote/google/repository"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/google"
)

type GoogleStorageBucketIamMemberEnumerator struct {
	repository        repository.AssetRepository
	storageRepository repository.StorageRepository
	factory           resource.ResourceFactory
}

func NewGoogleStorageBucketIamMemberEnumerator(repo repository.AssetRepository, storageRepo repository.StorageRepository, factory resource.ResourceFactory) *GoogleStorageBucketIamMemberEnumerator {
	return &GoogleStorageBucketIamMemberEnumerator{
		repository:        repo,
		storageRepository: storageRepo,
		factory:           factory,
	}
}

func (e *GoogleStorageBucketIamMemberEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleStorageBucketIamMemberResourceType
}

func (e *GoogleStorageBucketIamMemberEnumerator) Enumerate() ([]*resource.Resource, error) {
	resources, err := e.repository.SearchAllBuckets()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), google.GoogleStorageBucketResourceType)
	}

	results := make([]*resource.Resource, 0, len(resources))

	for _, bucket := range resources {
		bindings, err := e.storageRepository.ListAllBindings(bucket.DisplayName)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}
		for roleName, members := range bindings {
			for _, member := range members {
				id := fmt.Sprintf("b/%s/%s/%s", bucket.DisplayName, roleName, member)
				results = append(
					results,
					e.factory.CreateAbstractResource(
						string(e.SupportedType()),
						id,
						map[string]interface{}{
							"id":     id,
							"bucket": fmt.Sprintf("b/%s", bucket.DisplayName),
							"role":   roleName,
							"member": member,
						},
					),
				)
			}
		}
	}

	return results, err
}
