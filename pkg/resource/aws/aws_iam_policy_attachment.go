package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsIamPolicyAttachmentResourceType = "aws_iam_policy_attachment"

func initAwsIAMPolicyAttachmentMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsIamPolicyAttachmentResourceType, func(res *resource.AbstractResource) {
		val := res.Attrs
		val.SafeDelete([]string{"name"})
	})
}
