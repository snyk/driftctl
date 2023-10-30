package google

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const GoogleStorageBucketIamMemberResourceType = "google_storage_bucket_iam_member"

func initGoogleStorageBucketIamBMemberMetadata(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(GoogleStorageBucketIamMemberResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"force_destroy"})
		res.Attributes().SafeDelete([]string{"etag"})
	})
	resourceSchemaRepository.SetHumanReadableAttributesFunc(GoogleStorageBucketIamMemberResourceType, func(res *resource.Resource) map[string]string {
		attrs := map[string]string{
			"bucket": *res.Attrs.GetString("bucket"),
			"role":   *res.Attrs.GetString("role"),
			"member": *res.Attrs.GetString("member"),
		}
		return attrs
	})

}
