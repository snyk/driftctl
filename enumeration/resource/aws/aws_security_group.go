package aws

import "github.com/snyk/driftctl/enumeration/resource"

const AwsSecurityGroupResourceType = "aws_security_group"

func initAwsSecurityGroupMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsSecurityGroupResourceType, resource.FlagDeepMode)
}
