package aws

import (
	"context"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/client"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	"github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
)

func TestS3BucketAnalyticSupplier_Resources(t *testing.T) {

	tests := []struct {
		test      string
		dirName   string
		mocks     func(repository *repository.MockS3Repository)
		listError error
		wantErr   error
	}{
		{
			test:    "multiple bucket with multiple analytics",
			dirName: "s3_bucket_analytics_multiple",
			mocks: func(repository *repository.MockS3Repository) {
				repository.On(
					"ListAllBuckets",
				).Return([]*s3.Bucket{
					{Name: awssdk.String("bucket-martin-test-drift")},
					{Name: awssdk.String("bucket-martin-test-drift2")},
					{Name: awssdk.String("bucket-martin-test-drift3")},
				}, nil)

				repository.On(
					"GetBucketLocation",
					&s3.Bucket{Name: awssdk.String("bucket-martin-test-drift")},
				).Return(
					"eu-west-1",
					nil,
				)

				repository.On(
					"GetBucketLocation",
					&s3.Bucket{Name: awssdk.String("bucket-martin-test-drift2")},
				).Return(
					"eu-west-3",
					nil,
				)

				repository.On(
					"GetBucketLocation",
					&s3.Bucket{Name: awssdk.String("bucket-martin-test-drift3")},
				).Return(
					"ap-northeast-1",
					nil,
				)

				repository.On(
					"ListBucketAnalyticsConfigurations",
					&s3.Bucket{Name: awssdk.String("bucket-martin-test-drift")},
					"eu-west-1",
				).Return(
					[]*s3.AnalyticsConfiguration{
						{Id: awssdk.String("Analytics_Bucket1")},
						{Id: awssdk.String("Analytics2_Bucket1")},
					},
					nil,
				)

				repository.On(
					"ListBucketAnalyticsConfigurations",
					&s3.Bucket{Name: awssdk.String("bucket-martin-test-drift2")},
					"eu-west-3",
				).Return(
					[]*s3.AnalyticsConfiguration{
						{Id: awssdk.String("Analytics_Bucket2")},
						{Id: awssdk.String("Analytics2_Bucket2")},
					},
					nil,
				)

				repository.On(
					"ListBucketAnalyticsConfigurations",
					&s3.Bucket{Name: awssdk.String("bucket-martin-test-drift3")},
					"ap-northeast-1",
				).Return(
					[]*s3.AnalyticsConfiguration{
						{Id: awssdk.String("Analytics_Bucket3")},
						{Id: awssdk.String("Analytics2_Bucket3")},
					},
					nil,
				)
			},
		},

		{
			test: "cannot list bucket", dirName: "s3_bucket_analytics_list_bucket",
			mocks: func(repository *repository.MockS3Repository) {
				repository.On("ListAllBuckets").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: remoteerror.NewResourceEnumerationErrorWithType(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsS3BucketAnalyticsConfigurationResourceType, resourceaws.AwsS3BucketResourceType),
		},
		{
			test: "cannot list Analytics", dirName: "s3_bucket_analytics_list_analytics",
			mocks: func(repository *repository.MockS3Repository) {
				repository.On("ListAllBuckets").Return(
					[]*s3.Bucket{
						{Name: awssdk.String("bucket-martin-test-drift")},
					},
					nil,
				)
				repository.On(
					"GetBucketLocation",
					&s3.Bucket{Name: awssdk.String("bucket-martin-test-drift")},
				).Return(
					"eu-west-1",
					nil,
				)

				repository.On(
					"ListBucketAnalyticsConfigurations",
					&s3.Bucket{Name: awssdk.String("bucket-martin-test-drift")},
					"eu-west-1",
				).Return(
					nil,
					awserr.NewRequestFailure(nil, 403, ""),
				)

			},
			wantErr: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsS3BucketAnalyticsConfigurationResourceType),
		},
	}
	for _, tt := range tests {
		shouldUpdate := tt.dirName == *goldenfile.Update

		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		if shouldUpdate {
			provider, err := InitTestAwsProvider(providerLibrary)
			if err != nil {
				t.Fatal(err)
			}
			repository := repository.NewS3Repository(client.NewAWSClientFactory(provider.session))
			supplierLibrary.AddSupplier(NewS3BucketAnalyticSupplier(provider, repository))
		}

		t.Run(tt.test, func(t *testing.T) {

			mock := repository.MockS3Repository{}
			tt.mocks(&mock)

			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)

			deserializer := awsdeserializer.NewS3BucketAnalyticDeserializer()
			s := &S3BucketAnalyticSupplier{
				provider,
				deserializer,
				&mock,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(t, err, tt.wantErr)

			test.CtyTestDiff(got, tt.dirName, provider, deserializer, shouldUpdate, t)
		})
	}
}
