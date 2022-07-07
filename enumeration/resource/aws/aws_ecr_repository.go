package aws

import "github.com/snyk/driftctl/enumeration/resource"

const AwsEcrRepositoryResourceType = "aws_ecr_repository"

func initAwsEcrRepositoryMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {

	resourceSchemaRepository.SetFlags(AwsEcrRepositoryResourceType, resource.FlagDeepMode)
}
