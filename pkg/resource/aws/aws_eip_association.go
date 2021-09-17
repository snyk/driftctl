package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

const AwsEipAssociationResourceType = "aws_eip_association"

func initAwsEipAssociationMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsEipAssociationResourceType, resource.FlagDeepMode)
}
