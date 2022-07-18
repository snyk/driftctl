package google

import "github.com/snyk/driftctl/enumeration/resource"

const GoogleComputeNetworkResourceType = "google_compute_network"

func initGoogleComputeNetworkMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(GoogleComputeNetworkResourceType, resource.FlagDeepMode)
}
