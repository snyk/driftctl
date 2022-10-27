package repository

import (
	"fmt"

	"github.com/snyk/driftctl/enumeration/remote/aws/client"
	"github.com/snyk/driftctl/enumeration/remote/cache"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type S3Repository interface {
	ListAllBuckets() ([]*s3.Bucket, error)
	GetBucketNotification(bucketName, region string) (*s3.NotificationConfiguration, error)
	GetBucketPolicy(bucketName, region string) (*string, error)
	GetBucketPublicAccessBlock(bucketName, region string) (*s3.PublicAccessBlockConfiguration, error)
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
	cacheKey := "s3ListAllBuckets"
	v := s.cache.GetAndLock(cacheKey)
	defer s.cache.Unlock(cacheKey)
	if v != nil {
		return v.([]*s3.Bucket), nil
	}

	out, err := s.clientFactory.GetS3Client(nil).ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}
	s.cache.Put(cacheKey, out.Buckets)
	return out.Buckets, nil
}

func (s *s3Repository) GetBucketPolicy(bucketName, region string) (*string, error) {
	cacheKey := fmt.Sprintf("s3GetBucketPolicy_%s_%s", bucketName, region)
	if v := s.cache.Get(cacheKey); v != nil {
		return v.(*string), nil
	}
	policy, err := s.clientFactory.
		GetS3Client(&awssdk.Config{Region: &region}).
		GetBucketPolicy(
			&s3.GetBucketPolicyInput{Bucket: &bucketName},
		)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NoSuchBucketPolicy" {
				return nil, nil
			}
		}
		return nil, errors.Wrapf(
			err,
			"Error listing bucket policy %s",
			bucketName,
		)
	}

	result := policy.Policy
	if result != nil && *result == "" {
		result = nil
	}

	s.cache.Put(cacheKey, result)
	return result, nil
}

func (s *s3Repository) GetBucketPublicAccessBlock(bucketName, region string) (*s3.PublicAccessBlockConfiguration, error) {
	cacheKey := fmt.Sprintf("s3GetBucketPublicAccessBlock_%s_%s", bucketName, region)
	if v := s.cache.Get(cacheKey); v != nil {
		return v.(*s3.PublicAccessBlockConfiguration), nil
	}
	response, err := s.clientFactory.
		GetS3Client(&awssdk.Config{Region: &region}).
		GetPublicAccessBlock(&s3.GetPublicAccessBlockInput{Bucket: &bucketName})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NoSuchPublicAccessBlockConfiguration" {
				return nil, nil
			}
		}
		return nil, errors.Wrapf(
			err,
			"Error listing bucket public access block %s",
			bucketName,
		)
	}

	result := response.PublicAccessBlockConfiguration

	s.cache.Put(cacheKey, result)
	return result, nil
}

func (s *s3Repository) GetBucketNotification(bucketName, region string) (*s3.NotificationConfiguration, error) {
	cacheKey := fmt.Sprintf("s3GetBucketNotification_%s_%s", bucketName, region)
	if v := s.cache.Get(cacheKey); v != nil {
		return v.(*s3.NotificationConfiguration), nil
	}
	bucketNotificationConfig, err := s.clientFactory.
		GetS3Client(&awssdk.Config{Region: &region}).
		GetBucketNotificationConfiguration(
			&s3.GetBucketNotificationConfigurationRequest{Bucket: &bucketName},
		)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error listing bucket notification configuration %s",
			bucketName,
		)
	}

	result := bucketNotificationConfig
	if s.notificationIsEmpty(bucketNotificationConfig) {
		result = nil
	}

	s.cache.Put(cacheKey, result)
	return result, nil
}

func (s *s3Repository) notificationIsEmpty(notification *s3.NotificationConfiguration) bool {
	return notification.TopicConfigurations == nil &&
		notification.QueueConfigurations == nil &&
		notification.LambdaFunctionConfigurations == nil
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
	v := s.cache.GetAndLock(cacheKey)
	defer s.cache.Unlock(cacheKey)
	if v != nil {
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
