package google

import "github.com/snyk/driftctl/enumeration/resource"

const GoogleComputeRouterResourceType = "google_compute_router"

func initGoogleComputeRouterMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(GoogleComputeRouterResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"name":    *res.Attrs.GetString("name"),
			"region":  *res.Attrs.GetString("region"),
			"project": *res.Attrs.GetString("project"),
		}
	})
}
