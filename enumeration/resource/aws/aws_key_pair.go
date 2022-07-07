package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsKeyPairResourceType = "aws_key_pair"

func initAwsKeyPairMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsKeyPairResourceType, resource.FlagDeepMode)
}
