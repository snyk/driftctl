package google

import "github.com/cloudskiff/driftctl/pkg/resource"

const GoogleComputeDiskResourceType = "google_compute_disk"

func initGoogleComputeDiskMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(GoogleComputeDiskResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"Name": *res.Attributes().GetString("name"),
		}
	})
}
