package google

import "github.com/snyk/driftctl/enumeration/resource"

const GoogleStorageBucketResourceType = "google_storage_bucket"

func initGoogleStorageBucketMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(GoogleStorageBucketResourceType, resource.FlagDeepMode)
}
