package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type IamUserPolicyAttachmentDetailsFetcher struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
}

func NewIamUserPolicyAttachmentDetailsFetcher(provider terraform.ResourceReader, deserializer *resource.Deserializer) *IamUserPolicyAttachmentDetailsFetcher {
	return &IamUserPolicyAttachmentDetailsFetcher{
		reader:       provider,
		deserializer: deserializer,
	}
}

func (r *IamUserPolicyAttachmentDetailsFetcher) ReadDetails(res resource.Resource) (resource.Resource, error) {
	ctyVal, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: aws.AwsIamUserPolicyAttachmentResourceType,
		ID: res.TerraformId(),
		Attributes: map[string]string{
			"user":       *res.Attributes().GetString("user"),
			"policy_arn": *res.Attributes().GetString("policy_arn"),
		},
	})
	if err != nil {
		return nil, err
	}
	deserializedRes, err := r.deserializer.DeserializeOne(aws.AwsIamUserPolicyAttachmentResourceType, *ctyVal)
	if err != nil {
		return nil, err
	}

	return deserializedRes, nil
}
