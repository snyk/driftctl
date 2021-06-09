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
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	tf "github.com/cloudskiff/driftctl/pkg/remote/terraform"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	testresource "github.com/cloudskiff/driftctl/test/resource"

	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	"github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
)

func TestS3BucketMetricSupplier_Resources(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockS3Repository)
		wantErr error
	}{
		{
			test: "multiple bucket with multiple metrics", dirName: "s3_bucket_metrics_multiple",
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
					"ListBucketMetricsConfigurations",
					&s3.Bucket{Name: awssdk.String("bucket-martin-test-drift2")},
					"eu-west-3",
				).Return(
					[]*s3.MetricsConfiguration{
						{Id: awssdk.String("Metrics_Bucket2")},
						{Id: awssdk.String("Metrics2_Bucket2")},
					},
					nil,
				)
			},
		},
		{
			test: "cannot list bucket", dirName: "s3_bucket_metrics_list_bucket",
			mocks: func(repository *repository.MockS3Repository) {
				repository.On("ListAllBuckets").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: remoteerror.NewResourceEnumerationErrorWithType(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsS3BucketMetricResourceType, resourceaws.AwsS3BucketResourceType),
		},
		{
			test: "cannot list metrics", dirName: "s3_bucket_metrics_list_metrics",
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
					"eu-west-3",
					nil,
				)

				repository.On(
					"ListBucketMetricsConfigurations",
					&s3.Bucket{Name: awssdk.String("bucket-martin-test-drift")},
					"eu-west-3",
				).Return(
					nil,
					awserr.NewRequestFailure(nil, 403, ""),
				)

			},
			wantErr: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsS3BucketMetricResourceType),
		},
	}
	for _, tt := range tests {
		shouldUpdate := tt.dirName == *goldenfile.Update

		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		repo := testresource.InitFakeSchemaRepository("aws", "3.19.0")
		resourceaws.InitResourcesMetadata(repo)
		factory := terraform.NewTerraformResourceFactory(repo)

		deserializer := resource.NewDeserializer(factory)
		if shouldUpdate {
			provider, err := InitTestAwsProvider(providerLibrary)
			if err != nil {
				t.Fatal(err)
			}
			repository := repository.NewS3Repository(client.NewAWSClientFactory(provider.session), cache.New(0))
			supplierLibrary.AddSupplier(NewS3BucketMetricSupplier(provider, repository, deserializer))
		}

		t.Run(tt.test, func(t *testing.T) {

			mock := repository.MockS3Repository{}
			tt.mocks(&mock)

			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			s := &S3BucketMetricSupplier{
				provider,
				deserializer,
				&mock,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
				tf.TerraformProviderConfig{
					Name:         "test",
					DefaultAlias: "eu-west-3",
				},
			}
			got, err := s.Resources()
			assert.Equal(t, err, tt.wantErr)

			test.CtyTestDiff(got, tt.dirName, provider, deserializer, shouldUpdate, t)
		})
	}
}
