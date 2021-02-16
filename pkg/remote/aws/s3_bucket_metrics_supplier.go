package aws

import (
	"fmt"

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

type S3BucketMetricSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	repository   repository.S3Repository
	runner       *terraform.ParallelResourceReader
}

func NewS3BucketMetricSupplier(provider *AWSTerraformProvider, repository repository.S3Repository) *S3BucketMetricSupplier {
	return &S3BucketMetricSupplier{
		provider,
		awsdeserializer.NewS3BucketMetricDeserializer(),
		repository,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *S3BucketMetricSupplier) Resources() ([]resource.Resource, error) {
	buckets, err := s.repository.ListAllBuckets()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, aws.AwsS3BucketMetricResourceType, aws.AwsS3BucketResourceType)
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
		if err := s.listBucketMetricConfiguration(&bucket, region); err != nil {
			return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsS3BucketMetricResourceType)
		}
	}
	ctyVals, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(ctyVals)
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
