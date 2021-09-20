package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

const AwsDefaultVpcResourceType = "aws_default_vpc"

func initAwsDefaultVpcMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsDefaultVpcResourceType, resource.FlagDeepMode)
}
