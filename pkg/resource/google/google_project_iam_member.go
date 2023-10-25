package google

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const GoogleProjectIamMemberResourceType = "google_project_iam_member"

func initGoogleProjectIAMMemberMetadata(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(GoogleProjectIamMemberResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"force_destroy"})
		res.Attributes().SafeDelete([]string{"etag"})
	})
}
