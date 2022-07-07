package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsDefaultSecurityGroupResourceType = "aws_default_security_group"

func initAwsDefaultSecurityGroupMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsDefaultSecurityGroupResourceType, resource.FlagDeepMode)
}
