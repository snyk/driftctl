package repository

import (
	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/client"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type S3Repository interface {
	ListAllBuckets() ([]*s3.Bucket, error)
	ListBucketInventoryConfigurations(bucket *s3.Bucket, region string) ([]*s3.InventoryConfiguration, error)
	ListBucketMetricsConfigurations(bucket *s3.Bucket, region string) ([]*s3.MetricsConfiguration, error)
	ListBucketAnalyticsConfigurations(bucket *s3.Bucket, region string) ([]*s3.AnalyticsConfiguration, error)
	GetBucketLocation(bucket *s3.Bucket) (string, error)
}

type s3Repository struct {
	clientFactory client.AwsClientFactoryInterface
}

func NewS3Repository(factory client.AwsClientFactoryInterface) *s3Repository {
	return &s3Repository{
		factory,
	}
}

func (s *s3Repository) ListAllBuckets() ([]*s3.Bucket, error) {
	out, err := s.clientFactory.GetS3Client(nil).ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}
	return out.Buckets, nil
}

func (s *s3Repository) ListBucketInventoryConfigurations(bucket *s3.Bucket, region string) ([]*s3.InventoryConfiguration, error) {

	inventoryConfigurations := make([]*s3.InventoryConfiguration, 0)
	client := s.clientFactory.GetS3Client(&awssdk.Config{Region: &region})
	request := &s3.ListBucketInventoryConfigurationsInput{
		Bucket:            bucket.Name,
		ContinuationToken: nil,
	}

	for {
		configurations, err := client.ListBucketInventoryConfigurations(request)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"Error listing bucket inventory configuration %s",
				*bucket.Name,
			)
		}
		inventoryConfigurations = append(inventoryConfigurations, configurations.InventoryConfigurationList...)
		if configurations.IsTruncated != nil && *configurations.IsTruncated {
			request.ContinuationToken = configurations.NextContinuationToken
		} else {
			break
		}
	}

	return inventoryConfigurations, nil
}

func (s *s3Repository) ListBucketMetricsConfigurations(bucket *s3.Bucket, region string) ([]*s3.MetricsConfiguration, error) {
	metricsConfigurationList := make([]*s3.MetricsConfiguration, 0)
	client := s.clientFactory.GetS3Client(&awssdk.Config{Region: &region})
	request := &s3.ListBucketMetricsConfigurationsInput{
		Bucket:            bucket.Name,
		ContinuationToken: nil,
	}

	for {
		configurations, err := client.ListBucketMetricsConfigurations(request)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"Error listing bucket metrics configuration %s",
				*bucket.Name,
			)
		}
		metricsConfigurationList = append(metricsConfigurationList, configurations.MetricsConfigurationList...)
		if configurations.IsTruncated != nil && *configurations.IsTruncated {
			request.ContinuationToken = configurations.NextContinuationToken
		} else {
			break
		}
	}
	return metricsConfigurationList, nil
}

func (s *s3Repository) ListBucketAnalyticsConfigurations(bucket *s3.Bucket, region string) ([]*s3.AnalyticsConfiguration, error) {
	analyticsConfigurationList := make([]*s3.AnalyticsConfiguration, 0)
	client := s.clientFactory.GetS3Client(&awssdk.Config{Region: &region})
	request := &s3.ListBucketAnalyticsConfigurationsInput{
		Bucket:            bucket.Name,
		ContinuationToken: nil,
	}

	for {
		configurations, err := client.ListBucketAnalyticsConfigurations(request)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"Error listing bucket analytics configuration %s",
				*bucket.Name,
			)
		}
		analyticsConfigurationList = append(analyticsConfigurationList, configurations.AnalyticsConfigurationList...)

		if configurations.IsTruncated != nil && *configurations.IsTruncated {
			request.ContinuationToken = configurations.NextContinuationToken
		} else {
			break
		}
	}

	return analyticsConfigurationList, nil
}

func (s *s3Repository) GetBucketLocation(bucket *s3.Bucket) (string, error) {
	bucketLocationRequest := s3.GetBucketLocationInput{Bucket: bucket.Name}
	bucketLocationResponse, err := s.clientFactory.GetS3Client(nil).GetBucketLocation(&bucketLocationRequest)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == s3.ErrCodeNoSuchBucket {
			logrus.WithFields(logrus.Fields{
				"bucket": *bucket.Name,
			}).Warning("Unable to retrieve bucket region, this may be an inconsistency in S3 api for fresh deleted bucket, skipping ...")
			return "", nil
		}
		return "", err
	}

	// Buckets in Region us-east-1 have a LocationConstraint of null.
	// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetBucketLocation.html#API_GetBucketLocation_ResponseSyntax
	if bucketLocationResponse.LocationConstraint == nil {
		return "us-east-1", err
	}

	if *bucketLocationResponse.LocationConstraint == "EU" {
		return "eu-west-1", err
	}

	return *bucketLocationResponse.LocationConstraint, nil
}
