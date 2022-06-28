package google

import "github.com/snyk/driftctl/enumeration/resource"

const GoogleComputeNetworkResourceType = "google_compute_network"

func initGoogleComputeNetworkMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(GoogleComputeNetworkResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"name": *res.Attributes().GetString("name"),
		}
	})
	resourceSchemaRepository.SetFlags(GoogleComputeNetworkResourceType, resource.FlagDeepMode)
}
