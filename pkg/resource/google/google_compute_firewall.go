package google

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const GoogleComputeFirewallResourceType = "google_compute_firewall"

func initGoogleComputeFirewallMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(GoogleComputeFirewallResourceType, func(res *resource.Resource) {
		res.Attrs.SafeDelete([]string{"timeouts"})
	})
	resourceSchemaRepository.SetFlags(GoogleComputeFirewallResourceType, resource.FlagDeepMode)
}
