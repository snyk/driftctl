package aws

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type S3BucketInventoryDetailsFetcher struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
}

func NewS3BucketInventoryDetailsFetcher(provider terraform.ResourceReader, deserializer *resource.Deserializer) *S3BucketInventoryDetailsFetcher {
	return &S3BucketInventoryDetailsFetcher{
		reader:       provider,
		deserializer: deserializer,
	}
}

func (r *S3BucketInventoryDetailsFetcher) ReadDetails(res resource.Resource) (resource.Resource, error) {
	ctyVal, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: aws.AwsS3BucketInventoryResourceType,
		ID: res.TerraformId(),
		Attributes: map[string]string{
			"alias": *res.Attributes().GetString("region"),
		},
	})
	if err != nil {
		return nil, remoteerror.NewResourceScanningError(err, res.TerraformType())
	}
	deserializedRes, err := r.deserializer.DeserializeOne(aws.AwsS3BucketInventoryResourceType, *ctyVal)
	if err != nil {
		return nil, err
	}

	return deserializedRes, nil
}
