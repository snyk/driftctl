package google

import "github.com/snyk/driftctl/pkg/resource"

const GoogleComputeImageResourceType = "google_compute_image"

func initGoogleComputeImageMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(GoogleComputeImageResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"Name": *res.Attributes().GetString("name"),
		}
	})
}
