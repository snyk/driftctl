package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

const AwsVpcResourceType = "aws_vpc"

func initAwsVpcMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsVpcResourceType, resource.FlagDeepMode)
}
