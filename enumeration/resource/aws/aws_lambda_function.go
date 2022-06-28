package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsLambdaFunctionResourceType = "aws_lambda_function"

func initAwsLambdaFunctionMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {

	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsLambdaFunctionResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"function_name": res.ResourceId(),
		}
	})
	resourceSchemaRepository.SetFlags(AwsLambdaFunctionResourceType, resource.FlagDeepMode)
}
