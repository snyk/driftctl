package aws

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type S3BucketPolicyDetailsFetcher struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
}

func NewS3BucketPolicyDetailsFetcher(provider terraform.ResourceReader, deserializer *resource.Deserializer) *S3BucketPolicyDetailsFetcher {
	return &S3BucketPolicyDetailsFetcher{
		reader:       provider,
		deserializer: deserializer,
	}
}

func (r *S3BucketPolicyDetailsFetcher) ReadDetails(res resource.Resource) (resource.Resource, error) {
	ctyVal, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: aws.AwsS3BucketPolicyResourceType,
		ID: res.TerraformId(),
		Attributes: map[string]string{
			"alias": *res.Attributes().GetString("region"),
		},
	})
	if err != nil {
		return nil, remoteerror.NewResourceScanningError(err, res.TerraformType(), res.TerraformId())
	}
	deserializedRes, err := r.deserializer.DeserializeOne(aws.AwsS3BucketPolicyResourceType, *ctyVal)
	if err != nil {
		return nil, err
	}

	return deserializedRes, nil
}
