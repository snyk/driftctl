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

func TestS3BucketNotificationSupplier_Resources(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockS3Repository)
		wantErr error
	}{
		{
			test:    "single bucket without notifications",
			dirName: "s3_bucket_notifications_no_notif",
			mocks: func(repository *repository.MockS3Repository) {
				repository.On(
					"ListAllBuckets",
				).Return([]*s3.Bucket{
					{Name: awssdk.String("dritftctl-test-no-notifications")},
				}, nil)

				repository.On(
					"GetBucketLocation",
					&s3.Bucket{Name: awssdk.String("dritftctl-test-no-notifications")},
				).Return(
					"eu-west-3",
					nil,
				)
			},
		},
		{
			test: "multiple bucket with notifications", dirName: "s3_bucket_notifications_multiple",
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
			},
		},
		{
			test: "Cannot list bucket", dirName: "s3_bucket_notifications_list_bucket",
			mocks: func(repository *repository.MockS3Repository) {
				repository.On("ListAllBuckets").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: remoteerror.NewResourceEnumerationErrorWithType(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsS3BucketNotificationResourceType, resourceaws.AwsS3BucketResourceType),
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
			supplierLibrary.AddSupplier(NewS3BucketNotificationSupplier(provider, repository, deserializer))
		}

		t.Run(tt.test, func(t *testing.T) {

			mock := repository.MockS3Repository{}
			tt.mocks(&mock)

			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			s := &S3BucketNotificationSupplier{
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
