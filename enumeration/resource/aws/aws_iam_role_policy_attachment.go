package aws

import "github.com/snyk/driftctl/enumeration/resource"

const AwsIamRolePolicyAttachmentResourceType = "aws_iam_role_policy_attachment"

func initAwsIamRolePolicyAttachmentMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsIamRolePolicyAttachmentResourceType, resource.FlagDeepMode)
}
