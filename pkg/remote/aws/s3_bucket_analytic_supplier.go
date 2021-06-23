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

type S3BucketAnalyticSupplier struct {
	reader         terraform.ResourceReader
	deserializer   *resource.Deserializer
	repository     repository.S3Repository
	runner         *terraform.ParallelResourceReader
	providerConfig tf.TerraformProviderConfig
}

func NewS3BucketAnalyticSupplier(provider *AWSTerraformProvider, repository repository.S3Repository, deserializer *resource.Deserializer) *S3BucketAnalyticSupplier {
	return &S3BucketAnalyticSupplier{
		provider,
		deserializer,
		repository,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
		provider.Config,
	}
}

func (s *S3BucketAnalyticSupplier) Resources() ([]resource.Resource, error) {
	buckets, err := s.repository.ListAllBuckets()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, aws.AwsS3BucketAnalyticsConfigurationResourceType, aws.AwsS3BucketResourceType)
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
		if err := s.listBucketAnalyticConfiguration(&bucket, region); err != nil {
			return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsS3BucketAnalyticsConfigurationResourceType)
		}
	}
	ctyVals, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(aws.AwsS3BucketAnalyticsConfigurationResourceType, ctyVals)
}

func (s *S3BucketAnalyticSupplier) listBucketAnalyticConfiguration(bucket *s3.Bucket, region string) error {

	analyticsConfigurationList, err := s.repository.ListBucketAnalyticsConfigurations(bucket, region)
	if err != nil {
		return err
	}

	for _, analytics := range analyticsConfigurationList {
		id := fmt.Sprintf("%s:%s", *bucket.Name, *analytics.Id)
		s.runner.Run(func() (cty.Value, error) {
			s3BucketAnalytic, err := s.reader.ReadResource(
				terraform.ReadResourceArgs{
					Ty: aws.AwsS3BucketAnalyticsConfigurationResourceType,
					ID: id,
					Attributes: map[string]string{
						"alias": region,
					},
				},
			)

			if err != nil {
				return cty.NilVal, err
			}
			return *s3BucketAnalytic, err
		})

	}
	return nil
}
