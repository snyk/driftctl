package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsSubnetResourceType = "aws_subnet"

func initAwsSubnetMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsSubnetResourceType, resource.FlagDeepMode)
}
