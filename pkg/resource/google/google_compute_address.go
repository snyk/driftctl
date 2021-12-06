package google

import "github.com/snyk/driftctl/pkg/resource"

const GoogleComputeAddressResourceType = "google_compute_address"

func initGoogleComputeAddressMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(GoogleComputeAddressResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"Name":    *res.Attributes().GetString("name"),
			"Address": *res.Attributes().GetString("address"),
		}
	})
}
