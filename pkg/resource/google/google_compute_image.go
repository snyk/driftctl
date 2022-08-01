package google

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const GoogleComputeImageResourceType = "google_compute_image"

func initGoogleComputeImageMetadata(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(GoogleComputeImageResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"Name": *res.Attributes().GetString("name"),
		}
	})
}
