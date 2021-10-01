package google

import "github.com/cloudskiff/driftctl/pkg/resource"

const GoogleProjectIamMemberResourceType = "google_project_iam_member"

func initGoogleProjectIAMMemberMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(GoogleProjectIamMemberResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"force_destroy"})
		res.Attributes().SafeDelete([]string{"etag"})
	})
	resourceSchemaRepository.SetResolveReadAttributesFunc(GoogleProjectIamMemberResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"project": *res.Attrs.GetString("project"),
			"role":    *res.Attrs.GetString("role"),
			"member":  *res.Attrs.GetString("member"),
		}
	})
	resourceSchemaRepository.SetFlags(GoogleProjectIamMemberResourceType, resource.FlagDeepMode)

}
