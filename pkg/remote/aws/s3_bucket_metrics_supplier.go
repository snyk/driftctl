package aws

import (
	"fmt"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	awssdk "github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type S3BucketMetricSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	factory      AwsClientFactoryInterface
	runner       *terraform.ParallelResourceReader
}

func NewS3BucketMetricSupplier(provider *TerraformProvider, factory AwsClientFactoryInterface) *S3BucketMetricSupplier {
	return &S3BucketMetricSupplier{
		provider,
		awsdeserializer.NewS3BucketMetricDeserializer(),
		factory,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *S3BucketMetricSupplier) Resources() ([]resource.Resource, error) {
	input := &s3.ListBucketsInput{}

	client := s.factory.GetS3Client(nil)
	response, err := client.ListBuckets(input)
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, aws.AwsS3BucketMetricResourceType, aws.AwsS3BucketResourceType)
	}

	for _, bucket := range response.Buckets {
		name := *bucket.Name
		region, err := readBucketRegion(&client, name)
		if err != nil {
			return nil, err
		}
		if region == "" {
			continue
		}
		if err := s.listBucketMetricConfiguration(*bucket.Name, region); err != nil {
			return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsS3BucketMetricResourceType)
		}
	}
	ctyVals, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(ctyVals)
}

func (s *S3BucketMetricSupplier) listBucketMetricConfiguration(name, region string) error {
	request := &s3.ListBucketMetricsConfigurationsInput{
		Bucket:            &name,
		ContinuationToken: nil,
	}

	metricsConfigurationList := make([]*s3.MetricsConfiguration, 0)
	client := s.factory.GetS3Client(&awssdk.Config{Region: &region})

	for {
		configurations, err := client.ListBucketMetricsConfigurations(request)
		if err != nil {
			logrus.Warnf("Error listing bucket analytics configuration %s: %+v", name, err)
			return err
		}
		metricsConfigurationList = append(metricsConfigurationList, configurations.MetricsConfigurationList...)
		if configurations.IsTruncated != nil && *configurations.IsTruncated {
			request.ContinuationToken = configurations.NextContinuationToken
		} else {
			break
		}
	}

	for _, config := range metricsConfigurationList {
		id := fmt.Sprintf("%s:%s", name, *config.Id)
		s.runner.Run(func() (cty.Value, error) {
			s3BucketMetric, err := s.reader.ReadResource(
				terraform.ReadResourceArgs{
					Ty: aws.AwsS3BucketMetricResourceType,
					ID: id,
					Attributes: map[string]string{
						"aws_region": region,
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
