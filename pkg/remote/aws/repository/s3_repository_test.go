package repository

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/client"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	awstest "github.com/cloudskiff/driftctl/test/aws"
	"github.com/pkg/errors"
	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
)

func Test_s3Repository_ListAllBuckets(t *testing.T) {

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeS3)
		want    []*s3.Bucket
		wantErr error
	}{
		{
			name: "List buckets",
			mocks: func(client *awstest.MockFakeS3) {
				client.On("ListBuckets", &s3.ListBucketsInput{}).Return(
					&s3.ListBucketsOutput{
						Buckets: []*s3.Bucket{
							{Name: aws.String("bucket1")},
							{Name: aws.String("bucket2")},
							{Name: aws.String("bucket3")},
						},
					},
					nil,
				).Once()
			},
			want: []*s3.Bucket{
				{Name: aws.String("bucket1")},
				{Name: aws.String("bucket2")},
				{Name: aws.String("bucket3")},
			},
		},
		{
			name: "Error listing buckets",
			mocks: func(client *awstest.MockFakeS3) {
				client.On("ListBuckets", &s3.ListBucketsInput{}).Return(
					nil,
					awserr.NewRequestFailure(nil, 403, ""),
				).Once()
			},
			want:    nil,
			wantErr: awserr.NewRequestFailure(nil, 403, ""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			mockedClient := &awstest.MockFakeS3{}
			tt.mocks(mockedClient)
			factory := client.MockAwsClientFactoryInterface{}
			factory.On("GetS3Client", (*aws.Config)(nil)).Return(mockedClient).Once()
			r := NewS3Repository(&factory, store)
			got, err := r.ListAllBuckets()
			factory.AssertExpectations(t)
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllBuckets()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*s3.Bucket{}, store.Get("s3ListAllBuckets"))
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

func Test_s3Repository_ListBucketInventoryConfigurations(t *testing.T) {
	tests := []struct {
		name  string
		input struct {
			bucket s3.Bucket
			region string
		}
		mocks   func(client *awstest.MockFakeS3)
		want    []*s3.InventoryConfiguration
		wantErr string
	}{
		{
			name: "List inventory configs",
			input: struct {
				bucket s3.Bucket
				region string
			}{
				bucket: s3.Bucket{
					Name: awssdk.String("test-bucket"),
				},
				region: "us-east-1",
			},
			mocks: func(client *awstest.MockFakeS3) {
				client.On(
					"ListBucketInventoryConfigurations",
					&s3.ListBucketInventoryConfigurationsInput{
						Bucket:            awssdk.String("test-bucket"),
						ContinuationToken: nil,
					},
				).Return(
					&s3.ListBucketInventoryConfigurationsOutput{
						InventoryConfigurationList: []*s3.InventoryConfiguration{
							{Id: awssdk.String("config1")},
							{Id: awssdk.String("config2")},
							{Id: awssdk.String("config3")},
						},
						IsTruncated:           awssdk.Bool(true),
						NextContinuationToken: awssdk.String("nexttoken"),
					},
					nil,
				).Once()
				client.On(
					"ListBucketInventoryConfigurations",
					&s3.ListBucketInventoryConfigurationsInput{
						Bucket:            awssdk.String("test-bucket"),
						ContinuationToken: awssdk.String("nexttoken"),
					},
				).Return(
					&s3.ListBucketInventoryConfigurationsOutput{
						InventoryConfigurationList: []*s3.InventoryConfiguration{
							{Id: awssdk.String("config4")},
							{Id: awssdk.String("config5")},
							{Id: awssdk.String("config6")},
						},
						IsTruncated: awssdk.Bool(false),
					},
					nil,
				).Once()
			},
			want: []*s3.InventoryConfiguration{
				{Id: awssdk.String("config1")},
				{Id: awssdk.String("config2")},
				{Id: awssdk.String("config3")},
				{Id: awssdk.String("config4")},
				{Id: awssdk.String("config5")},
				{Id: awssdk.String("config6")},
			},
		},
		{
			name: "Error listing inventory configs",
			input: struct {
				bucket s3.Bucket
				region string
			}{
				bucket: s3.Bucket{
					Name: awssdk.String("test-bucket"),
				},
				region: "us-east-1",
			},
			mocks: func(client *awstest.MockFakeS3) {
				client.On(
					"ListBucketInventoryConfigurations",
					&s3.ListBucketInventoryConfigurationsInput{
						Bucket: awssdk.String("test-bucket"),
					},
				).Return(
					nil,
					errors.New("aws error"),
				).Once()
			},
			want:    nil,
			wantErr: "Error listing bucket inventory configuration test-bucket: aws error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			mockedClient := &awstest.MockFakeS3{}
			tt.mocks(mockedClient)
			factory := client.MockAwsClientFactoryInterface{}
			factory.On("GetS3Client", &aws.Config{Region: awssdk.String(tt.input.region)}).Return(mockedClient).Once()
			r := NewS3Repository(&factory, store)
			got, err := r.ListBucketInventoryConfigurations(&tt.input.bucket, tt.input.region)
			factory.AssertExpectations(t)
			if err != nil && tt.wantErr == "" {
				t.Fatalf("Unexpected error %+v", err)
			}
			if err != nil {
				assert.Equal(t, tt.wantErr, err.Error())
			}

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListBucketInventoryConfigurations(&tt.input.bucket, tt.input.region)
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*s3.InventoryConfiguration{}, store.Get(fmt.Sprintf("s3ListBucketInventoryConfigurations_%s_%s", *tt.input.bucket.Name, tt.input.region)))
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

func Test_s3Repository_ListBucketMetricsConfigurations(t *testing.T) {
	tests := []struct {
		name  string
		input struct {
			bucket s3.Bucket
			region string
		}
		mocks   func(client *awstest.MockFakeS3)
		want    []*s3.MetricsConfiguration
		wantErr string
	}{
		{
			name: "List metrics configs",
			input: struct {
				bucket s3.Bucket
				region string
			}{
				bucket: s3.Bucket{
					Name: awssdk.String("test-bucket"),
				},
				region: "us-east-1",
			},
			mocks: func(client *awstest.MockFakeS3) {
				client.On(
					"ListBucketMetricsConfigurations",
					&s3.ListBucketMetricsConfigurationsInput{
						Bucket:            awssdk.String("test-bucket"),
						ContinuationToken: nil,
					},
				).Return(
					&s3.ListBucketMetricsConfigurationsOutput{
						MetricsConfigurationList: []*s3.MetricsConfiguration{
							{Id: awssdk.String("metric1")},
							{Id: awssdk.String("metric2")},
							{Id: awssdk.String("metric3")},
						},
						IsTruncated:           awssdk.Bool(true),
						NextContinuationToken: awssdk.String("nexttoken"),
					},
					nil,
				).Once()
				client.On(
					"ListBucketMetricsConfigurations",
					&s3.ListBucketMetricsConfigurationsInput{
						Bucket:            awssdk.String("test-bucket"),
						ContinuationToken: awssdk.String("nexttoken"),
					},
				).Return(
					&s3.ListBucketMetricsConfigurationsOutput{
						MetricsConfigurationList: []*s3.MetricsConfiguration{
							{Id: awssdk.String("metric4")},
							{Id: awssdk.String("metric5")},
							{Id: awssdk.String("metric6")},
						},
						IsTruncated: awssdk.Bool(false),
					},
					nil,
				).Once()
			},
			want: []*s3.MetricsConfiguration{
				{Id: awssdk.String("metric1")},
				{Id: awssdk.String("metric2")},
				{Id: awssdk.String("metric3")},
				{Id: awssdk.String("metric4")},
				{Id: awssdk.String("metric5")},
				{Id: awssdk.String("metric6")},
			},
		},
		{
			name: "Error listing metrics configs",
			input: struct {
				bucket s3.Bucket
				region string
			}{
				bucket: s3.Bucket{
					Name: awssdk.String("test-bucket"),
				},
				region: "us-east-1",
			},
			mocks: func(client *awstest.MockFakeS3) {
				client.On(
					"ListBucketMetricsConfigurations",
					&s3.ListBucketMetricsConfigurationsInput{
						Bucket: awssdk.String("test-bucket"),
					},
				).Return(
					nil,
					errors.New("aws error"),
				).Once()
			},
			want:    nil,
			wantErr: "Error listing bucket metrics configuration test-bucket: aws error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			mockedClient := &awstest.MockFakeS3{}
			tt.mocks(mockedClient)
			factory := client.MockAwsClientFactoryInterface{}
			factory.On("GetS3Client", &aws.Config{Region: awssdk.String(tt.input.region)}).Return(mockedClient).Once()
			r := NewS3Repository(&factory, store)
			got, err := r.ListBucketMetricsConfigurations(&tt.input.bucket, tt.input.region)
			factory.AssertExpectations(t)
			if err != nil && tt.wantErr == "" {
				t.Fatalf("Unexpected error %+v", err)
			}
			if err != nil {
				assert.Equal(t, tt.wantErr, err.Error())
			}

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListBucketMetricsConfigurations(&tt.input.bucket, tt.input.region)
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*s3.MetricsConfiguration{}, store.Get(fmt.Sprintf("s3ListBucketMetricsConfigurations_%s_%s", *tt.input.bucket.Name, tt.input.region)))
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

func Test_s3Repository_ListBucketAnalyticsConfigurations(t *testing.T) {
	tests := []struct {
		name  string
		input struct {
			bucket s3.Bucket
			region string
		}
		mocks   func(client *awstest.MockFakeS3)
		want    []*s3.AnalyticsConfiguration
		wantErr string
	}{
		{
			name: "List analytics configs",
			input: struct {
				bucket s3.Bucket
				region string
			}{
				bucket: s3.Bucket{
					Name: awssdk.String("test-bucket"),
				},
				region: "us-east-1",
			},
			mocks: func(client *awstest.MockFakeS3) {
				client.On(
					"ListBucketAnalyticsConfigurations",
					&s3.ListBucketAnalyticsConfigurationsInput{
						Bucket:            awssdk.String("test-bucket"),
						ContinuationToken: nil,
					},
				).Return(
					&s3.ListBucketAnalyticsConfigurationsOutput{
						AnalyticsConfigurationList: []*s3.AnalyticsConfiguration{
							{Id: awssdk.String("analytic1")},
							{Id: awssdk.String("analytic2")},
							{Id: awssdk.String("analytic3")},
						},
						IsTruncated:           awssdk.Bool(true),
						NextContinuationToken: awssdk.String("nexttoken"),
					},
					nil,
				).Once()
				client.On(
					"ListBucketAnalyticsConfigurations",
					&s3.ListBucketAnalyticsConfigurationsInput{
						Bucket:            awssdk.String("test-bucket"),
						ContinuationToken: awssdk.String("nexttoken"),
					},
				).Return(
					&s3.ListBucketAnalyticsConfigurationsOutput{
						AnalyticsConfigurationList: []*s3.AnalyticsConfiguration{
							{Id: awssdk.String("analytic4")},
							{Id: awssdk.String("analytic5")},
							{Id: awssdk.String("analytic6")},
						},
						IsTruncated: awssdk.Bool(false),
					},
					nil,
				).Once()
			},
			want: []*s3.AnalyticsConfiguration{
				{Id: awssdk.String("analytic1")},
				{Id: awssdk.String("analytic2")},
				{Id: awssdk.String("analytic3")},
				{Id: awssdk.String("analytic4")},
				{Id: awssdk.String("analytic5")},
				{Id: awssdk.String("analytic6")},
			},
		},
		{
			name: "Error listing analytics configs",
			input: struct {
				bucket s3.Bucket
				region string
			}{
				bucket: s3.Bucket{
					Name: awssdk.String("test-bucket"),
				},
				region: "us-east-1",
			},
			mocks: func(client *awstest.MockFakeS3) {
				client.On(
					"ListBucketAnalyticsConfigurations",
					&s3.ListBucketAnalyticsConfigurationsInput{
						Bucket: awssdk.String("test-bucket"),
					},
				).Return(
					nil,
					errors.New("aws error"),
				).Once()
			},
			want:    nil,
			wantErr: "Error listing bucket analytics configuration test-bucket: aws error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			mockedClient := &awstest.MockFakeS3{}
			tt.mocks(mockedClient)
			factory := client.MockAwsClientFactoryInterface{}
			factory.On("GetS3Client", &aws.Config{Region: awssdk.String(tt.input.region)}).Return(mockedClient).Once()
			r := NewS3Repository(&factory, store)
			got, err := r.ListBucketAnalyticsConfigurations(&tt.input.bucket, tt.input.region)
			factory.AssertExpectations(t)
			if err != nil && tt.wantErr == "" {
				t.Fatalf("Unexpected error %+v", err)
			}
			if err != nil {
				assert.Equal(t, tt.wantErr, err.Error())
			}

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListBucketAnalyticsConfigurations(&tt.input.bucket, tt.input.region)
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*s3.AnalyticsConfiguration{}, store.Get(fmt.Sprintf("s3ListBucketAnalyticsConfigurations_%s_%s", *tt.input.bucket.Name, tt.input.region)))
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

func Test_s3Repository_GetBucketLocation(t *testing.T) {

	tests := []struct {
		name    string
		bucket  *s3.Bucket
		mocks   func(client *awstest.MockFakeS3)
		want    string
		wantErr string
	}{
		{
			name: "get bucket location",
			bucket: &s3.Bucket{
				Name: awssdk.String("test-bucket"),
			},
			mocks: func(client *awstest.MockFakeS3) {
				client.On("GetBucketLocation", &s3.GetBucketLocationInput{
					Bucket: awssdk.String("test-bucket"),
				}).Return(
					&s3.GetBucketLocationOutput{
						LocationConstraint: awssdk.String("eu-east-1"),
					},
					nil,
				).Once()
			},
			want: "eu-east-1",
		},
		{
			name: "get bucket location for us-east-2",
			bucket: &s3.Bucket{
				Name: awssdk.String("test-bucket"),
			},
			mocks: func(client *awstest.MockFakeS3) {
				client.On("GetBucketLocation", &s3.GetBucketLocationInput{
					Bucket: awssdk.String("test-bucket"),
				}).Return(
					&s3.GetBucketLocationOutput{},
					nil,
				).Once()
			},
			want: "us-east-1",
		},
		{
			name: "get bucket location when no such bucket",
			bucket: &s3.Bucket{
				Name: awssdk.String("test-bucket"),
			},
			mocks: func(client *awstest.MockFakeS3) {
				client.On("GetBucketLocation", &s3.GetBucketLocationInput{
					Bucket: awssdk.String("test-bucket"),
				}).Return(
					nil,
					awserr.New(s3.ErrCodeNoSuchBucket, "", nil),
				).Once()
			},
			want: "",
		},
		{
			name: "get bucket location when error",
			bucket: &s3.Bucket{
				Name: awssdk.String("test-bucket"),
			},
			mocks: func(client *awstest.MockFakeS3) {
				client.On("GetBucketLocation", &s3.GetBucketLocationInput{
					Bucket: awssdk.String("test-bucket"),
				}).Return(
					nil,
					awserr.New("UnknownError", "aws error", nil),
				).Once()
			},
			wantErr: "UnknownError: aws error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			mockedClient := &awstest.MockFakeS3{}
			tt.mocks(mockedClient)
			factory := client.MockAwsClientFactoryInterface{}
			factory.On("GetS3Client", (*aws.Config)(nil)).Return(mockedClient).Once()
			r := NewS3Repository(&factory, store)
			got, err := r.GetBucketLocation(*tt.bucket.Name)
			factory.AssertExpectations(t)
			if err != nil && tt.wantErr == "" {
				t.Fatalf("Unexpected error %+v", err)
			}
			if err != nil {
				assert.Equal(t, tt.wantErr, err.Error())
			}

			if err == nil && tt.want != "" {
				// Check that results were cached
				cachedData, err := r.GetBucketLocation(*tt.bucket.Name)
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, "", store.Get(fmt.Sprintf("s3GetBucketLocation_%s", *tt.bucket.Name)))
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}
