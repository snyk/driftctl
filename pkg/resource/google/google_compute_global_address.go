package google

import "github.com/cloudskiff/driftctl/pkg/resource"

const GoogleComputeGlobalAddressResourceType = "google_compute_global_address"

func initGoogleComputeGlobalAddressMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(GoogleComputeGlobalAddressResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"Name":    *res.Attributes().GetString("name"),
			"Address": *res.Attributes().GetString("address"),
		}
	})
}
