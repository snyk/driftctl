package google

import "github.com/snyk/driftctl/enumeration/resource"

const GoogleProjectIamMemberResourceType = "google_project_iam_member"

func initGoogleProjectIAMMemberMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(GoogleProjectIamMemberResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"project": *res.Attrs.GetString("project"),
			"role":    *res.Attrs.GetString("role"),
			"member":  *res.Attrs.GetString("member"),
		}
	})
	resourceSchemaRepository.SetFlags(GoogleProjectIamMemberResourceType, resource.FlagDeepMode)

}
