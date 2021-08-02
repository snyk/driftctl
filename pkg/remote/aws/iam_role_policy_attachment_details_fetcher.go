package aws

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type IamRolePolicyAttachmentDetailsFetcher struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
}

func NewIamRolePolicyAttachmentDetailsFetcher(provider terraform.ResourceReader, deserializer *resource.Deserializer) *IamRolePolicyAttachmentDetailsFetcher {
	return &IamRolePolicyAttachmentDetailsFetcher{
		reader:       provider,
		deserializer: deserializer,
	}
}

func (r *IamRolePolicyAttachmentDetailsFetcher) ReadDetails(res resource.Resource) (resource.Resource, error) {
	ctyVal, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: aws.AwsIamRolePolicyAttachmentResourceType,
		ID: res.TerraformId(),
		Attributes: map[string]string{
			"role":       *res.Attributes().GetString("role"),
			"policy_arn": *res.Attributes().GetString("policy_arn"),
		},
	})
	if err != nil {
		return nil, remoteerror.NewResourceScanningError(err, res.TerraformType())
	}
	deserializedRes, err := r.deserializer.DeserializeOne(aws.AwsIamRolePolicyAttachmentResourceType, *ctyVal)
	if err != nil {
		return nil, err
	}

	return deserializedRes, nil
}
