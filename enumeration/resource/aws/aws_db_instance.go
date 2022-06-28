package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsDbInstanceResourceType = "aws_db_instance"

func initAwsDbInstanceMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsDbInstanceResourceType, resource.FlagDeepMode)
}
