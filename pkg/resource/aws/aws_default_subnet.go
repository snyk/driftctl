package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsDefaultSubnetResourceType = "aws_default_subnet"

func initAwsDefaultSubnetMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsDefaultSubnetResourceType, func(res *resource.AbstractResource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
}
