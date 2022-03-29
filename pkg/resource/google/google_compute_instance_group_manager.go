package google

import "github.com/snyk/driftctl/pkg/resource"

const GoogleComputeInstanceGroupManagerResourceType = "google_compute_instance_group_manager"

func initComputeInstanceGroupManagerMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(GoogleComputeInstanceGroupManagerResourceType, func(res *resource.Resource) map[string]string {
		attrs := make(map[string]string)
		if v := res.Attributes().GetString("name"); v != nil && *v != "" {
			attrs["Name"] = *v
		}
		return attrs
	})
}
