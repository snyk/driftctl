package google

import "github.com/cloudskiff/driftctl/pkg/resource"

const GoogleComputeFirewallResourceType = "google_compute_firewall"

func initGoogleComputeFirewallMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(GoogleComputeFirewallResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"name":    *res.Attrs.GetString("name"),
			"project": *res.Attrs.GetString("project"),
		}
	})
	resourceSchemaRepository.SetNormalizeFunc(GoogleComputeFirewallResourceType, func(res *resource.Resource) {
		res.Attrs.SafeDelete([]string{"timeouts"})
	})
}
