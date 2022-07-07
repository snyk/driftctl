package aws

import "github.com/snyk/driftctl/enumeration/resource"

const AwsVpcResourceType = "aws_vpc"

func initAwsVpcMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsVpcResourceType, resource.FlagDeepMode)
}
