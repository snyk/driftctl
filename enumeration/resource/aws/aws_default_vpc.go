package aws

import "github.com/snyk/driftctl/enumeration/resource"

const AwsDefaultVpcResourceType = "aws_default_vpc"

func initAwsDefaultVpcMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsDefaultVpcResourceType, resource.FlagDeepMode)
}
