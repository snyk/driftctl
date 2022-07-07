package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsIamPolicyAttachmentResourceType = "aws_iam_policy_attachment"

func initAwsIAMPolicyAttachmentMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsIamPolicyAttachmentResourceType, resource.FlagDeepMode)
}
