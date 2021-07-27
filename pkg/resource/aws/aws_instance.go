package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/hashicorp/go-version"
)

const AwsInstanceResourceType = "aws_instance"

func initAwsInstanceMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsInstanceResourceType, func(res *resource.AbstractResource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})

		if v, _ := version.NewVersion("3.38.0"); res.Schema().ProviderVersion.LessThan(v) {
			val.SafeDelete([]string{"instance_initiated_shutdown_behavior"})
		}
	})
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsInstanceResourceType, func(res *resource.AbstractResource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if tags := val.GetMap("tags"); tags != nil {
			if name, ok := tags["Name"]; ok {
				attrs["Name"] = name.(string)
			}
		}
		return attrs
	})
}
