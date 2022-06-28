package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsDynamodbTableResourceType = "aws_dynamodb_table"

func initAwsDynamodbTableMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsDynamodbTableResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"table_name": res.ResourceId(),
		}
	})
	resourceSchemaRepository.SetFlags(AwsDynamodbTableResourceType, resource.FlagDeepMode)

}
