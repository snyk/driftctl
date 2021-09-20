package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

const AwsIamUserPolicyAttachmentResourceType = "aws_iam_user_policy_attachment"

func initAwsIamUserPolicyAttachmentMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsIamUserPolicyAttachmentResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"user":       *res.Attributes().GetString("user"),
			"policy_arn": *res.Attributes().GetString("policy_arn"),
		}
	})
	resourceSchemaRepository.SetFlags(AwsIamUserPolicyAttachmentResourceType, resource.FlagDeepMode)
}
