package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	tf "github.com/cloudskiff/driftctl/pkg/remote/terraform"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/zclconf/go-cty/cty"
)

type S3BucketMetricSupplier struct {
	reader         terraform.ResourceReader
	deserializer   *resource.Deserializer
	repository     repository.S3Repository
	runner         *terraform.ParallelResourceReader
	providerConfig tf.TerraformProviderConfig
}

func NewS3BucketMetricSupplier(provider *AWSTerraformProvider, repository repository.S3Repository, deserializer *resource.Deserializer) *S3BucketMetricSupplier {
	return &S3BucketMetricSupplier{
		provider,
		deserializer,
		repository,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
		provider.Config,
	}
}

func (s *S3BucketMetricSupplier) Resources() ([]resource.Resource, error) {
	buckets, err := s.repository.ListAllBuckets()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, aws.AwsS3BucketMetricResourceType, aws.AwsS3BucketResourceType)
	}

	for _, bucket := range buckets {
		bucket := *bucket
		region, err := s.repository.GetBucketLocation(*bucket.Name)
		if err != nil {
			return nil, err
		}
		if region == "" || region != s.providerConfig.DefaultAlias {
			continue
		}
		if err := s.listBucketMetricConfiguration(&bucket, region); err != nil {
			return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsS3BucketMetricResourceType)
		}
	}
	ctyVals, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(aws.AwsS3BucketMetricResourceType, ctyVals)
}

func (s *S3BucketMetricSupplier) listBucketMetricConfiguration(bucket *s3.Bucket, region string) error {

	metricsConfigurationList, err := s.repository.ListBucketMetricsConfigurations(bucket, region)
	if err != nil {
		return err
	}

	for _, config := range metricsConfigurationList {
		id := fmt.Sprintf("%s:%s", *bucket.Name, *config.Id)
		s.runner.Run(func() (cty.Value, error) {
			s3BucketMetric, err := s.reader.ReadResource(
				terraform.ReadResourceArgs{
					Ty: aws.AwsS3BucketMetricResourceType,
					ID: id,
					Attributes: map[string]string{
						"alias": region,
					},
				},
			)
			if err != nil {
				return cty.NilVal, err
			}
			return *s3BucketMetric, err
		})
	}
	return nil
}
