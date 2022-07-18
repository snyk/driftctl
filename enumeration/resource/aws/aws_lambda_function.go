package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsLambdaFunctionResourceType = "aws_lambda_function"

func initAwsLambdaFunctionMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsLambdaFunctionResourceType, resource.FlagDeepMode)
}
