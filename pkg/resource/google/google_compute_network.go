package google

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const GoogleComputeNetworkResourceType = "google_compute_network"

func initGoogleComputeNetworkMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(GoogleComputeNetworkResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"timeouts"})
		res.Attributes().SafeDelete([]string{"self_link"})
		res.Attributes().SafeDelete([]string{"gateway_ipv4"})
		res.Attributes().SafeDelete([]string{"delete_default_routes_on_create"})
	})
	resourceSchemaRepository.SetFlags(GoogleComputeNetworkResourceType, resource.FlagDeepMode)
}
