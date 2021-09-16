package aws

import (
	"strconv"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/hashicorp/terraform/flatmap"
)

const AwsCloudformationStackResourceType = "aws_cloudformation_stack"

func initAwsCloudformationStackMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsCloudformationStackResourceType, func(res *resource.Resource) map[string]string {
		attrs := make(map[string]interface{})
		if v := res.Attributes().GetMap("parameters"); v != nil {
			attrs["parameters.%"] = strconv.FormatInt(int64(len(v)), 10)
			attrs["parameters"] = v
		}
		return flatmap.Flatten(attrs)
	})
	resourceSchemaRepository.SetNormalizeFunc(AwsCloudformationStackResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
}
