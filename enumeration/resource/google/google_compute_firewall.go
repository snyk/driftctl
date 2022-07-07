package google

import "github.com/snyk/driftctl/enumeration/resource"

const GoogleComputeFirewallResourceType = "google_compute_firewall"

func initGoogleComputeFirewallMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(GoogleComputeFirewallResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"name":    *res.Attrs.GetString("name"),
			"project": *res.Attrs.GetString("project"),
		}
	})
	resourceSchemaRepository.SetFlags(GoogleComputeFirewallResourceType, resource.FlagDeepMode)
}
