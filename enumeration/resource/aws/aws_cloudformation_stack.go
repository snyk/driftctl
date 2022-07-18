package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsCloudformationStackResourceType = "aws_cloudformation_stack"

func initAwsCloudformationStackMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsCloudformationStackResourceType, resource.FlagDeepMode)
}
