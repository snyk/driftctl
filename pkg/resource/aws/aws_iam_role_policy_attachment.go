package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsIamRolePolicyAttachmentResourceType = "aws_iam_role_policy_attachment"

func initAwsIamRolePolicyAttachmentMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsIamRolePolicyAttachmentResourceType, resource.FlagDeepMode)
}
