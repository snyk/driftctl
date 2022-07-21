package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsIamUserPolicyAttachmentResourceType = "aws_iam_user_policy_attachment"

func initAwsIamUserPolicyAttachmentMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsIamUserPolicyAttachmentResourceType, resource.FlagDeepMode)
}
