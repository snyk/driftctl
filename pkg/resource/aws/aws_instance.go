package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsInstanceResourceType = "aws_instance"

func initAwsInstanceMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsInstanceResourceType, func(res *resource.AbstractResource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
		val.SafeDelete([]string{"instance_initiated_shutdown_behavior"})
	})
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsInstanceResourceType, func(res *resource.AbstractResource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if tags := val.GetStringMap("tags"); tags != nil {
			if name, ok := tags["name"]; ok {
				attrs["Name"] = name
			}
		}
		return attrs
	})
}
