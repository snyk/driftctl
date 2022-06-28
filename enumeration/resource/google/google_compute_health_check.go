package google

import "github.com/snyk/driftctl/enumeration/resource"

const GoogleComputeHealthCheckResourceType = "google_compute_health_check"

func initGoogleComputeHealthCheckMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(GoogleComputeHealthCheckResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"Name": *res.Attributes().GetString("name"),
		}
	})
}
