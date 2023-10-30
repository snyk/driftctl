package google

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const GoogleStorageBucketResourceType = "google_storage_bucket"

func initGoogleStorageBucketMetadata(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(GoogleStorageBucketResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"force_destroy"})
	})
}
