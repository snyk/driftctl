package google

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const GoogleProjectIamMemberResourceType = "google_project_iam_member"

func initGoogleProjectIAMMemberMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(GoogleProjectIamMemberResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"force_destroy"})
		res.Attributes().SafeDelete([]string{"etag"})
	})
	resourceSchemaRepository.SetFlags(GoogleProjectIamMemberResourceType, resource.FlagDeepMode)
}
