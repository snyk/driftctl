package aws

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/zclconf/go-cty/cty"
)

type S3BucketSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	repository   repository.S3Repository
	runner       *terraform.ParallelResourceReader
}

func NewS3BucketSupplier(provider *AWSTerraformProvider, repository repository.S3Repository) *S3BucketSupplier {
	return &S3BucketSupplier{
		provider,
		awsdeserializer.NewS3BucketDeserializer(),
		repository,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s S3BucketSupplier) Resources() ([]resource.Resource, error) {
	buckets, err := s.repository.ListAllBuckets()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsS3BucketResourceType)
	}

	for _, bucket := range buckets {
		b := *bucket
		s.runner.Run(func() (cty.Value, error) {
			return s.readBucket(b)
		})
	}
	values, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(values)
}

func (s *S3BucketSupplier) readBucket(bucket s3.Bucket) (cty.Value, error) {
	region, err := s.repository.GetBucketLocation(&bucket)
	if err != nil {
		return cty.NilVal, err
	}
	if region == "" {
		return cty.NilVal, nil
	}

	s3Bucket, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: aws.AwsS3BucketResourceType,
		ID: *bucket.Name,
		Attributes: map[string]string{
			"alias": region,
		},
	})
	if err != nil {
		return cty.NilVal, err
	}
	return *s3Bucket, err
}
