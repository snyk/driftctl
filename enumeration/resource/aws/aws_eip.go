package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsEipResourceType = "aws_eip"

func initAwsEipMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsEipResourceType, resource.FlagDeepMode)
}
