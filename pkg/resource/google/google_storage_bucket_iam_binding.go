package google

import "github.com/cloudskiff/driftctl/pkg/resource"

const GoogleStorageBucketIamBindingResourceType = "google_storage_bucket_iam_binding"

func initGoogleStorageBucketIamBindingMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(GoogleStorageBucketIamBindingResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"force_destroy"})
		res.Attributes().SafeDelete([]string{"etag"})
	})
	resourceSchemaRepository.SetResolveReadAttributesFunc(GoogleStorageBucketIamBindingResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"bucket": *res.Attrs.GetString("bucket"),
			"role":   *res.Attrs.GetString("role"),
		}
	})
	resourceSchemaRepository.SetFlags(GoogleStorageBucketIamBindingResourceType, resource.FlagDeepMode)

}
