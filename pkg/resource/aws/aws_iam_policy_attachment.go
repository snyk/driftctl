package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsIamPolicyAttachmentResourceType = "aws_iam_policy_attachment"

func initAwsIAMPolicyAttachmentMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsIamPolicyAttachmentResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"name"})
	})
	resourceSchemaRepository.SetFlags(AwsIamPolicyAttachmentResourceType, resource.FlagDeepMode)
}
