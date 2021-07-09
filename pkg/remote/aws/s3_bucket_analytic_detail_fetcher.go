package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type S3BucketAnalyticDetailsFetcher struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
}

func NewS3BucketAnalyticDetailsFetcher(provider terraform.ResourceReader, deserializer *resource.Deserializer) *S3BucketAnalyticDetailsFetcher {
	return &S3BucketAnalyticDetailsFetcher{
		reader:       provider,
		deserializer: deserializer,
	}
}

func (r *S3BucketAnalyticDetailsFetcher) ReadDetails(res resource.Resource) (resource.Resource, error) {
	ctyVal, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: aws.AwsS3BucketAnalyticsConfigurationResourceType,
		ID: res.TerraformId(),
		Attributes: map[string]string{
			"alias": *res.Attributes().GetString("region"),
		},
	})
	if err != nil {
		return nil, err
	}
	deserializedRes, err := r.deserializer.DeserializeOne(aws.AwsS3BucketAnalyticsConfigurationResourceType, *ctyVal)
	if err != nil {
		return nil, err
	}

	return deserializedRes, nil
}
