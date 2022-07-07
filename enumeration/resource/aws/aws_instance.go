package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsInstanceResourceType = "aws_instance"

func initAwsInstanceMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsInstanceResourceType, func(res *resource.Resource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if tags := val.GetMap("tags"); tags != nil {
			if name, ok := tags["Name"]; ok {
				attrs["Name"] = name.(string)
			}
		}
		return attrs
	})
	resourceSchemaRepository.SetFlags(AwsInstanceResourceType, resource.FlagDeepMode)
}
