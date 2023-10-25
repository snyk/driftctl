package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsLambdaFunctionResourceType = "aws_lambda_function"

func initAwsLambdaFunctionMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsLambdaFunctionResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
		val.SafeDelete([]string{"publish"})
		val.SafeDelete([]string{"last_modified"})
		val.SafeDelete([]string{"filename"})
		val.DeleteIfDefault("code_signing_config_arn")
		val.DeleteIfDefault("image_uri")
		val.DeleteIfDefault("package_type")
		val.DeleteIfDefault("signing_job_arn")
		val.DeleteIfDefault("signing_profile_version_arn")
		val.SafeDelete([]string{"source_code_size"})
	})
}
