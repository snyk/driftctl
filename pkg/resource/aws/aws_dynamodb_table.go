package aws

import "github.com/snyk/driftctl/enumeration/resource"

const AwsDynamodbTableResourceType = "aws_dynamodb_table"

func initAwsDynamodbTableMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsDynamodbTableResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
	resourceSchemaRepository.SetFlags(AwsDynamodbTableResourceType, resource.FlagDeepMode)
}
