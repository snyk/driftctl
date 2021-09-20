package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsDbSubnetGroupResourceType = "aws_db_subnet_group"

func initAwsDbSubnetGroupMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsDbSubnetGroupResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"name_prefix"})
	})
	resourceSchemaRepository.SetFlags(AwsDbSubnetGroupResourceType, resource.FlagDeepMode)
}
