package aws

import (
	"github.com/snyk/driftctl/pkg/resource"
)

const AwsDefaultSubnetResourceType = "aws_default_subnet"

func initAwsDefaultSubnetMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsDefaultSubnetResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
	resourceSchemaRepository.SetFlags(AwsDefaultSubnetResourceType, resource.FlagDeepMode)
}
