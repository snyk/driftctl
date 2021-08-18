package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsDynamodbTableResourceType = "aws_dynamodb_table"

func initAwsDynamodbTableMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsDynamodbTableResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"table_name": res.ResourceId(),
		}
	})
	resourceSchemaRepository.SetNormalizeFunc(AwsDynamodbTableResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
}
