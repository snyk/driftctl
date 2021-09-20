package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsSubnetResourceType = "aws_subnet"

func initAwsSubnetMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsSubnetResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
	resourceSchemaRepository.SetFlags(AwsSubnetResourceType, resource.FlagDeepMode)
}
