package aws

import "github.com/snyk/driftctl/pkg/resource"

const AwsIamRolePolicyAttachmentResourceType = "aws_iam_role_policy_attachment"

func initAwsIamRolePolicyAttachmentMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsIamRolePolicyAttachmentResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"role":       *res.Attributes().GetString("role"),
			"policy_arn": *res.Attributes().GetString("policy_arn"),
		}
	})
	resourceSchemaRepository.SetFlags(AwsIamRolePolicyAttachmentResourceType, resource.FlagDeepMode)
}
