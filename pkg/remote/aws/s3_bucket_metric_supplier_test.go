package aws

import (
	"context"
	"testing"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/stretchr/testify/assert"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/mocks"
)

func TestS3BucketMetricSupplier_Resources(t *testing.T) {

	tests := []struct {
		test           string
		dirName        string
		bucketsIDs     []string
		bucketLocation map[string]string
		metricsIDs     map[string][]string
		listError      error
		wantErr        error
	}{
		{
			test: "multiple bucket with multiple metrics", dirName: "s3_bucket_metrics_multiple",
			bucketsIDs: []string{
				"bucket-martin-test-drift",
				"bucket-martin-test-drift2",
				"bucket-martin-test-drift3",
			},
			bucketLocation: map[string]string{
				"bucket-martin-test-drift":  "eu-west-1",
				"bucket-martin-test-drift2": "eu-west-3",
				"bucket-martin-test-drift3": "ap-northeast-1",
			},
			metricsIDs: map[string][]string{
				"bucket-martin-test-drift": {
					"Metrics_Bucket1",
					"Metrics2_Bucket1",
				},
				"bucket-martin-test-drift2": {
					"Metrics_Bucket2",
					"Metrics2_Bucket2",
				},
				"bucket-martin-test-drift3": {
					"Metrics_Bucket3",
					"Metrics2_Bucket3",
				},
			},
		},
		{
			test: "cannot list bucket", dirName: "s3_bucket_metrics_list_bucket",
			bucketsIDs: nil,
			listError:  awserr.NewRequestFailure(nil, 403, ""),
			bucketLocation: map[string]string{
				"bucket-martin-test-drift":  "eu-west-1",
				"bucket-martin-test-drift2": "eu-west-3",
				"bucket-martin-test-drift3": "ap-northeast-1",
			},
			metricsIDs: map[string][]string{
				"bucket-martin-test-drift": {
					"Metrics_Bucket1",
					"Metrics2_Bucket1",
				},
				"bucket-martin-test-drift2": {
					"Metrics_Bucket2",
					"Metrics2_Bucket2",
				},
				"bucket-martin-test-drift3": {
					"Metrics_Bucket3",
					"Metrics2_Bucket3",
				},
			},
			wantErr: remoteerror.NewResourceEnumerationErrorWithType(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsS3BucketMetricResourceType, resourceaws.AwsS3BucketResourceType),
		},
		{
			test: "cannot list metrics", dirName: "s3_bucket_metrics_list_metrics",
			bucketsIDs: []string{
				"bucket-martin-test-drift",
				"bucket-martin-test-drift2",
				"bucket-martin-test-drift3",
			},
			bucketLocation: map[string]string{
				"bucket-martin-test-drift":  "eu-west-1",
				"bucket-martin-test-drift2": "eu-west-3",
				"bucket-martin-test-drift3": "ap-northeast-1",
			},
			metricsIDs: nil,
			listError:  awserr.NewRequestFailure(nil, 403, ""),
			wantErr:    remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsS3BucketMetricResourceType),
		},
	}
	for _, tt := range tests {
		shouldUpdate := tt.dirName == *goldenfile.Update

		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		if shouldUpdate {
			provider, err := NewTerraFormProvider()
			if err != nil {
				t.Fatal(err)
			}

			factory := AwsClientFactory{config: provider.session}

			providerLibrary.AddProvider(terraform.AWS, provider)
			supplierLibrary.AddSupplier(NewS3BucketMetricSupplier(provider, factory))
		}

		t.Run(tt.test, func(t *testing.T) {

			mock := mocks.NewMockAWSS3Client(tt.bucketsIDs, nil, nil, tt.metricsIDs, tt.bucketLocation, tt.listError)
			factory := mocks.NewMockAwsClientFactory(mock)

			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewS3BucketMetricDeserializer()
			s := &S3BucketMetricSupplier{
				provider,
				deserializer,
				factory,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(t, err, tt.wantErr)

			test.CtyTestDiff(got, tt.dirName, provider, deserializer, shouldUpdate, t)
		})
	}
}
