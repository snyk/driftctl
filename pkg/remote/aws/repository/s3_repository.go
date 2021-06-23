package repository

import (
	"fmt"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/client"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type S3Repository interface {
	ListAllBuckets() ([]*s3.Bucket, error)
	ListBucketInventoryConfigurations(bucket *s3.Bucket, region string) ([]*s3.InventoryConfiguration, error)
	ListBucketMetricsConfigurations(bucket *s3.Bucket, region string) ([]*s3.MetricsConfiguration, error)
	ListBucketAnalyticsConfigurations(bucket *s3.Bucket, region string) ([]*s3.AnalyticsConfiguration, error)
	GetBucketLocation(bucketName string) (string, error)
}

type s3Repository struct {
	clientFactory client.AwsClientFactoryInterface
	cache         cache.Cache
}

func NewS3Repository(factory client.AwsClientFactoryInterface, c cache.Cache) *s3Repository {
	return &s3Repository{
		factory,
		c,
	}
}

func (s *s3Repository) ListAllBuckets() ([]*s3.Bucket, error) {
	if v := s.cache.Get("s3ListAllBuckets"); v != nil {
		return v.([]*s3.Bucket), nil
	}

	out, err := s.clientFactory.GetS3Client(nil).ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}
	s.cache.Put("s3ListAllBuckets", out.Buckets)
	return out.Buckets, nil
}

func (s *s3Repository) ListBucketInventoryConfigurations(bucket *s3.Bucket, region string) ([]*s3.InventoryConfiguration, error) {
	cacheKey := fmt.Sprintf("s3ListBucketInventoryConfigurations_%s_%s", *bucket.Name, region)
	if v := s.cache.Get(cacheKey); v != nil {
		return v.([]*s3.InventoryConfiguration), nil
	}

	inventoryConfigurations := make([]*s3.InventoryConfiguration, 0)
	s3client := s.clientFactory.GetS3Client(&awssdk.Config{Region: &region})
	request := &s3.ListBucketInventoryConfigurationsInput{
		Bucket:            bucket.Name,
		ContinuationToken: nil,
	}

	for {
		configurations, err := s3client.ListBucketInventoryConfigurations(request)
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

	s.cache.Put(cacheKey, inventoryConfigurations)
	return inventoryConfigurations, nil
}

func (s *s3Repository) ListBucketMetricsConfigurations(bucket *s3.Bucket, region string) ([]*s3.MetricsConfiguration, error) {
	cacheKey := fmt.Sprintf("s3ListBucketMetricsConfigurations_%s_%s", *bucket.Name, region)
	if v := s.cache.Get(cacheKey); v != nil {
		return v.([]*s3.MetricsConfiguration), nil
	}

	metricsConfigurationList := make([]*s3.MetricsConfiguration, 0)
	s3client := s.clientFactory.GetS3Client(&awssdk.Config{Region: &region})
	request := &s3.ListBucketMetricsConfigurationsInput{
		Bucket:            bucket.Name,
		ContinuationToken: nil,
	}

	for {
		configurations, err := s3client.ListBucketMetricsConfigurations(request)
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

	s.cache.Put(cacheKey, metricsConfigurationList)
	return metricsConfigurationList, nil
}

func (s *s3Repository) ListBucketAnalyticsConfigurations(bucket *s3.Bucket, region string) ([]*s3.AnalyticsConfiguration, error) {
	cacheKey := fmt.Sprintf("s3ListBucketAnalyticsConfigurations_%s_%s", *bucket.Name, region)
	if v := s.cache.Get(cacheKey); v != nil {
		return v.([]*s3.AnalyticsConfiguration), nil
	}

	analyticsConfigurationList := make([]*s3.AnalyticsConfiguration, 0)
	s3client := s.clientFactory.GetS3Client(&awssdk.Config{Region: &region})
	request := &s3.ListBucketAnalyticsConfigurationsInput{
		Bucket:            bucket.Name,
		ContinuationToken: nil,
	}

	for {
		configurations, err := s3client.ListBucketAnalyticsConfigurations(request)
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

	s.cache.Put(cacheKey, analyticsConfigurationList)
	return analyticsConfigurationList, nil
}

func (s *s3Repository) GetBucketLocation(bucketName string) (string, error) {
	cacheKey := fmt.Sprintf("s3GetBucketLocation_%s", bucketName)
	if v := s.cache.Get(cacheKey); v != nil {
		return v.(string), nil
	}

	bucketLocationRequest := s3.GetBucketLocationInput{Bucket: &bucketName}
	bucketLocationResponse, err := s.clientFactory.GetS3Client(nil).GetBucketLocation(&bucketLocationRequest)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == s3.ErrCodeNoSuchBucket {
			logrus.WithFields(logrus.Fields{
				"bucket": bucketName,
			}).Warning("Unable to retrieve bucket region, this may be an inconsistency in S3 api for fresh deleted bucket, skipping ...")
			return "", nil
		}
		return "", err
	}

	var location string

	// Buckets in Region us-east-1 have a LocationConstraint of null.
	// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetBucketLocation.html#API_GetBucketLocation_ResponseSyntax
	if bucketLocationResponse.LocationConstraint == nil {
		location = "us-east-1"
	} else {
		location = *bucketLocationResponse.LocationConstraint
	}

	if location == "EU" {
		location = "eu-west-1"
	}

	s.cache.Put(cacheKey, location)
	return location, nil
}
