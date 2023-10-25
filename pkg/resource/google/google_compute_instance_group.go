package google

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const GoogleComputeInstanceGroupResourceType = "google_compute_instance_group"

func initGoogleComputeInstanceGroupMetadata(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(GoogleComputeInstanceGroupResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"timeouts"})
	})
	resourceSchemaRepository.SetHumanReadableAttributesFunc(GoogleComputeInstanceGroupResourceType, func(res *resource.Resource) map[string]string {
		attrs := make(map[string]string)
		if v := res.Attributes().GetString("name"); v != nil && *v != "" {
			attrs["Name"] = *v
		}
		return attrs
	})
}
