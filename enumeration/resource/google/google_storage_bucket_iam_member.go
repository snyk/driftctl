package google

import "github.com/snyk/driftctl/enumeration/resource"

const GoogleStorageBucketIamMemberResourceType = "google_storage_bucket_iam_member"

func initGoogleStorageBucketIamBMemberMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(GoogleStorageBucketIamMemberResourceType, func(res *resource.Resource) map[string]string {
		attrs := map[string]string{
			"bucket": *res.Attrs.GetString("bucket"),
			"role":   *res.Attrs.GetString("role"),
			"member": *res.Attrs.GetString("member"),
		}
		return attrs
	})
	resourceSchemaRepository.SetFlags(GoogleStorageBucketIamMemberResourceType, resource.FlagDeepMode)

}
