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

type S3BucketAnalyticSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	factory      AwsClientFactoryInterface
	runner       *terraform.ParallelResourceReader
}

func NewS3BucketAnalyticSupplier(provider *TerraformProvider, factory AwsClientFactoryInterface) *S3BucketAnalyticSupplier {
	return &S3BucketAnalyticSupplier{
		provider,
		awsdeserializer.NewS3BucketAnalyticDeserializer(),
		factory,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *S3BucketAnalyticSupplier) Resources() ([]resource.Resource, error) {
	input := &s3.ListBucketsInput{}

	client := s.factory.GetS3Client(nil)
	response, err := client.ListBuckets(input)
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, aws.AwsS3BucketAnalyticsConfigurationResourceType, aws.AwsS3BucketResourceType)
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
		if err := s.listBucketAnalyticConfiguration(*bucket.Name, region); err != nil {
			return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsS3BucketAnalyticsConfigurationResourceType)
		}
	}
	ctyVals, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(ctyVals)
}

func (s *S3BucketAnalyticSupplier) listBucketAnalyticConfiguration(name, region string) error {
	request := &s3.ListBucketAnalyticsConfigurationsInput{
		Bucket:            &name,
		ContinuationToken: nil,
	}
	analyticsConfigurationList := make([]*s3.AnalyticsConfiguration, 0)
	client := s.factory.GetS3Client(&awssdk.Config{Region: &region})

	for {
		configurations, err := client.ListBucketAnalyticsConfigurations(request)
		if err != nil {
			logrus.Warnf("Error listing bucket analytics configuration %s: %+v", name, err)
			return err
		}
		analyticsConfigurationList = append(analyticsConfigurationList, configurations.AnalyticsConfigurationList...)

		if configurations.IsTruncated != nil && *configurations.IsTruncated {
			request.ContinuationToken = configurations.NextContinuationToken
		} else {
			break
		}
	}

	for _, analytics := range analyticsConfigurationList {
		id := fmt.Sprintf("%s:%s", name, *analytics.Id)
		s.runner.Run(func() (cty.Value, error) {
			s3BucketAnalytic, err := s.reader.ReadResource(
				terraform.ReadResourceArgs{
					Ty: aws.AwsS3BucketAnalyticsConfigurationResourceType,
					ID: id,
					Attributes: map[string]string{
						"aws_region": region,
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
