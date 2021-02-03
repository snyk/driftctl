package mocks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type MockAWSS3Client struct {
	s3iface.S3API
	bucketsIDs      []string
	analyticsIDs    map[string][]string
	inventoriesIDs  map[string][]string
	metricsIDs      map[string][]string
	bucketLocations map[string]string
	err             error
}

func NewMockAWSS3Client(bucketsIDs []string, analyticsIDs map[string][]string, inventoriesIDs map[string][]string, metricsIDs map[string][]string, bucketLocations map[string]string, err error) *MockAWSS3Client {
	return &MockAWSS3Client{bucketsIDs: bucketsIDs, analyticsIDs: analyticsIDs, inventoriesIDs: inventoriesIDs, metricsIDs: metricsIDs, bucketLocations: bucketLocations, err: err}
}

func (m MockAWSS3Client) ListBucketAnalyticsConfigurations(in *s3.ListBucketAnalyticsConfigurationsInput) (*s3.ListBucketAnalyticsConfigurationsOutput, error) {
	if m.analyticsIDs == nil && m.err != nil {
		return nil, m.err
	}

	var configurations []*s3.AnalyticsConfiguration
	for _, id := range m.analyticsIDs[*in.Bucket] {
		configurations = append(configurations, &s3.AnalyticsConfiguration{
			Id: aws.String(id),
		})
	}
	return &s3.ListBucketAnalyticsConfigurationsOutput{
		AnalyticsConfigurationList: configurations,
	}, nil
}
func (m MockAWSS3Client) ListBucketInventoryConfigurations(in *s3.ListBucketInventoryConfigurationsInput) (*s3.ListBucketInventoryConfigurationsOutput, error) {
	if m.inventoriesIDs == nil && m.err != nil {
		return nil, m.err
	}

	var configurations []*s3.InventoryConfiguration
	for _, id := range m.inventoriesIDs[*in.Bucket] {
		configurations = append(configurations, &s3.InventoryConfiguration{
			Id: aws.String(id),
		})
	}
	return &s3.ListBucketInventoryConfigurationsOutput{
		InventoryConfigurationList: configurations,
	}, nil
}
func (m MockAWSS3Client) ListBucketMetricsConfigurations(in *s3.ListBucketMetricsConfigurationsInput) (*s3.ListBucketMetricsConfigurationsOutput, error) {
	if m.metricsIDs == nil && m.err != nil {
		return nil, m.err
	}

	var configurations []*s3.MetricsConfiguration
	for _, id := range m.metricsIDs[*in.Bucket] {
		configurations = append(configurations, &s3.MetricsConfiguration{
			Id: aws.String(id),
		})
	}
	return &s3.ListBucketMetricsConfigurationsOutput{
		MetricsConfigurationList: configurations,
	}, nil
}

func (m MockAWSS3Client) ListBuckets(*s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
	if m.bucketsIDs == nil && m.err != nil {
		return nil, m.err
	}

	var buckets []*s3.Bucket
	for _, id := range m.bucketsIDs {
		buckets = append(buckets, &s3.Bucket{
			Name: aws.String(id),
		})
	}
	return &s3.ListBucketsOutput{
		Buckets: buckets,
		Owner:   nil,
	}, nil
}

func (m MockAWSS3Client) GetBucketLocation(input *s3.GetBucketLocationInput) (*s3.GetBucketLocationOutput, error) {
	location, exists := m.bucketLocations[*input.Bucket]
	if !exists {
		panic(fmt.Sprintf("no region provided for bucket %s", *input.Bucket))
	}
	return &s3.GetBucketLocationOutput{
		LocationConstraint: &location,
	}, nil
}
