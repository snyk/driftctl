package google

import "github.com/snyk/driftctl/enumeration/resource"

const GoogleProjectIamMemberResourceType = "google_project_iam_member"

func initGoogleProjectIAMMemberMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(GoogleProjectIamMemberResourceType, resource.FlagDeepMode)

}
