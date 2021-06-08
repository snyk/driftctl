package aws

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/zclconf/go-cty/cty"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	tf "github.com/cloudskiff/driftctl/pkg/remote/terraform"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type S3BucketSupplier struct {
	reader         terraform.ResourceReader
	deserializer   *resource.Deserializer
	repository     repository.S3Repository
	runner         *terraform.ParallelResourceReader
	providerConfig tf.TerraformProviderConfig
}

func NewS3BucketSupplier(provider *AWSTerraformProvider, repository repository.S3Repository, deserializer *resource.Deserializer) *S3BucketSupplier {
	return &S3BucketSupplier{
		provider,
		deserializer,
		repository,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
		provider.Config,
	}
}

func (s *S3BucketSupplier) SuppliedType() resource.ResourceType {
	return aws.AwsS3BucketResourceType
}

func (s *S3BucketSupplier) Resources() ([]resource.Resource, error) {
	buckets, err := s.repository.ListAllBuckets()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
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

	return s.deserializer.Deserialize(s.SuppliedType(), values)
}

func (s *S3BucketSupplier) readBucket(bucket s3.Bucket) (cty.Value, error) {
	region, err := s.repository.GetBucketLocation(&bucket)
	if err != nil {
		return cty.NilVal, err
	}
	if region == "" || region != s.providerConfig.DefaultAlias {
		return cty.NilVal, nil
	}

	s3Bucket, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: s.SuppliedType(),
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
