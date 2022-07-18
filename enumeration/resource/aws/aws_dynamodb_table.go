package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsDynamodbTableResourceType = "aws_dynamodb_table"

func initAwsDynamodbTableMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsDynamodbTableResourceType, resource.FlagDeepMode)

}
