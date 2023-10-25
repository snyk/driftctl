package google

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const GoogleComputeFirewallResourceType = "google_compute_firewall"

func initGoogleComputeFirewallMetadata(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(GoogleComputeFirewallResourceType, func(res *resource.Resource) {
		res.Attrs.SafeDelete([]string{"timeouts"})
	})
}
