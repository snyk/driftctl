package aws

import "github.com/snyk/driftctl/enumeration/resource"

const AwsIamUserPolicyAttachmentResourceType = "aws_iam_user_policy_attachment"

func initAwsIamUserPolicyAttachmentMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsIamUserPolicyAttachmentResourceType, resource.FlagDeepMode)
}
