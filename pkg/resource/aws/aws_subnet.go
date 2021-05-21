package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsSubnetResourceType = "aws_subnet"

func initAwsSubnetMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsSubnetResourceType, func(res *resource.AbstractResource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
}
