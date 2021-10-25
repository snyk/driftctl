package google

import "github.com/cloudskiff/driftctl/pkg/resource"

const GoogleComputeSubnetworkResourceType = "google_compute_subnetwork"

func initGoogleComputeSubnetworkMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(GoogleComputeSubnetworkResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"name":   *res.Attributes().GetString("display_name"),
			"region": *res.Attributes().GetString("location"),
		}
	})
	resourceSchemaRepository.SetNormalizeFunc(GoogleComputeSubnetworkResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"timeouts"})
		res.Attributes().SafeDelete([]string{"self_link"})
	})
	resourceSchemaRepository.SetHumanReadableAttributesFunc(GoogleComputeSubnetworkResourceType, func(res *resource.Resource) map[string]string {
		attrs := make(map[string]string)

		if v := res.Attributes().GetString("display_name"); v != nil && *v != "" {
			attrs["Name"] = *v
		}
		return attrs
	})
	resourceSchemaRepository.SetFlags(GoogleComputeSubnetworkResourceType, resource.FlagDeepMode)
}
