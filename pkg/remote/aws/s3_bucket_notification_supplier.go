package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/zclconf/go-cty/cty"
)

type S3BucketNotificationSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	repository   repository.S3Repository
	runner       *terraform.ParallelResourceReader
}

func NewS3BucketNotificationSupplier(provider *AWSTerraformProvider, repository repository.S3Repository) *S3BucketNotificationSupplier {
	return &S3BucketNotificationSupplier{
		provider,
		awsdeserializer.NewS3BucketNotificationDeserializer(),
		repository,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *S3BucketNotificationSupplier) Resources() ([]resource.Resource, error) {
	buckets, err := s.repository.ListAllBuckets()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, aws.AwsS3BucketNotificationResourceType, aws.AwsS3BucketResourceType)
	}

	for _, bucket := range buckets {
		bucket := *bucket
		region, err := s.repository.GetBucketLocation(&bucket)
		if err != nil {
			return nil, err
		}
		if region == "" {
			continue
		}
		s.runner.Run(func() (cty.Value, error) {
			s3BucketPolicy, err := s.reader.ReadResource(terraform.ReadResourceArgs{
				Ty: aws.AwsS3BucketNotificationResourceType,
				ID: *bucket.Name,
				Attributes: map[string]string{
					"alias": region,
				},
			})
			if err != nil {
				return cty.NilVal, err
			}
			return *s3BucketPolicy, err
		})
	}
	ctyVals, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}
	deserializedValues, err := s.deserializer.Deserialize(ctyVals)
	results := make([]resource.Resource, 0, len(deserializedValues))
	if err != nil {
		return deserializedValues, err
	}
	for _, val := range deserializedValues {
		res, ok := val.(*aws.AwsS3BucketNotification)
		if ok {
			if (res.LambdaFunction != nil && len(*res.LambdaFunction) > 0) ||
				(res.Queue != nil && len(*res.Queue) > 0) ||
				(res.Topic != nil && len(*res.Topic) > 0) {
				results = append(results, res)
			}
		}
	}
	return results, nil
}
