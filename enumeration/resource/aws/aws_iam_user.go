package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsIamUserResourceType = "aws_iam_user"

func initAwsIAMUserMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsIamUserResourceType, resource.FlagDeepMode)
}
