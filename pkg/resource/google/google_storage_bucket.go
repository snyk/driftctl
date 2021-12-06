package google

import "github.com/snyk/driftctl/pkg/resource"

const GoogleStorageBucketResourceType = "google_storage_bucket"

func initGoogleStorageBucketMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(GoogleStorageBucketResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"force_destroy"})
	})
	resourceSchemaRepository.SetResolveReadAttributesFunc(GoogleStorageBucketResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"name": res.ResourceId(),
		}
	})
	resourceSchemaRepository.SetFlags(GoogleStorageBucketResourceType, resource.FlagDeepMode)
}
