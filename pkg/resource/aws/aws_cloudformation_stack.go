package aws

import "github.com/snyk/driftctl/enumeration/resource"

const AwsCloudformationStackResourceType = "aws_cloudformation_stack"

func initAwsCloudformationStackMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsCloudformationStackResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
	resourceSchemaRepository.SetFlags(AwsCloudformationStackResourceType, resource.FlagDeepMode)
}
