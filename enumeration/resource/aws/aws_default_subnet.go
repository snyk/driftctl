package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsDefaultSubnetResourceType = "aws_default_subnet"

func initAwsDefaultSubnetMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsDefaultSubnetResourceType, resource.FlagDeepMode)
}
