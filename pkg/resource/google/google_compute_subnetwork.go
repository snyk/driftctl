package google

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const GoogleComputeSubnetworkResourceType = "google_compute_subnetwork"

func initGoogleComputeSubnetworkMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(GoogleComputeSubnetworkResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"timeouts"})
		res.Attributes().SafeDelete([]string{"self_link"})
	})
	resourceSchemaRepository.SetHumanReadableAttributesFunc(GoogleComputeSubnetworkResourceType, func(res *resource.Resource) map[string]string {
		attrs := make(map[string]string)

		if v := res.Attributes().GetString("name"); v != nil && *v != "" {
			attrs["Name"] = *v
		}
		return attrs
	})
	resourceSchemaRepository.SetFlags(GoogleComputeSubnetworkResourceType, resource.FlagDeepMode)
}
