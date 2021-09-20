package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

const AwsEcrRepositoryResourceType = "aws_ecr_repository"

func initAwsEcrRepositoryMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsEcrRepositoryResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
	resourceSchemaRepository.SetFlags(AwsEcrRepositoryResourceType, resource.FlagDeepMode)
}
