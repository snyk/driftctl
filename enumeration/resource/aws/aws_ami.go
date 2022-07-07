package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsAmiResourceType = "aws_ami"

func initAwsAmiMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsAmiResourceType, resource.FlagDeepMode)
}
