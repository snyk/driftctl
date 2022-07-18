package google

import "github.com/snyk/driftctl/enumeration/resource"

const GoogleComputeFirewallResourceType = "google_compute_firewall"

func initGoogleComputeFirewallMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(GoogleComputeFirewallResourceType, resource.FlagDeepMode)
}
