package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

const AwsEcrRepositoryResourceType = "aws_ecr_repository"

func initAwsEcrRepositoryMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsEcrRepositoryResourceType, func(res *resource.AbstractResource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
}
