package google

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const GoogleStorageBucketResourceType = "google_storage_bucket"

func initGoogleStorageBucketMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(GoogleStorageBucketResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"force_destroy"})
	})
	resourceSchemaRepository.SetFlags(GoogleStorageBucketResourceType, resource.FlagDeepMode)
}
