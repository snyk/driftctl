package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsKmsAliasResourceType = "aws_kms_alias"

func initAwsKmsAliasMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsKmsAliasResourceType, resource.FlagDeepMode)
}
