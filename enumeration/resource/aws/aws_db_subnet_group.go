package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsDbSubnetGroupResourceType = "aws_db_subnet_group"

func initAwsDbSubnetGroupMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsDbSubnetGroupResourceType, resource.FlagDeepMode)
}
