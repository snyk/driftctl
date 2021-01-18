package aws

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/parallel"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/mocks"
)

func TestS3BucketAnalyticSupplier_Resources(t *testing.T) {

	tests := []struct {
		test           string
		dirName        string
		bucketsIDs     []string
		bucketLocation map[string]string
		analyticsIDs   map[string][]string
		wantErr        bool
	}{
		{
			test: "multiple bucket with multiple analytics", dirName: "s3_bucket_analytics_multiple",
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
			analyticsIDs: map[string][]string{
				"bucket-martin-test-drift": {
					"Analytics_Bucket1",
					"Analytics2_Bucket1",
				},
				"bucket-martin-test-drift2": {
					"Analytics_Bucket2",
					"Analytics2_Bucket2",
				},
				"bucket-martin-test-drift3": {
					"Analytics_Bucket3",
					"Analytics2_Bucket3",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		shouldUpdate := tt.dirName == *goldenfile.Update
		if shouldUpdate {
			provider, err := NewTerraFormProvider()
			if err != nil {
				t.Fatal(err)
			}

			factory := AwsClientFactory{config: provider.session}
			terraform.AddProvider(terraform.AWS, provider)
			resource.AddSupplier(NewS3BucketAnalyticSupplier(provider.Runner().SubRunner(), factory))
		}

		t.Run(tt.test, func(t *testing.T) {

			mock := mocks.NewMockAWSS3Client(tt.bucketsIDs, tt.analyticsIDs, nil, nil, tt.bucketLocation)

			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, terraform.Provider(terraform.AWS), shouldUpdate)
			factory := mocks.NewMockAwsClientFactory(mock)

			deserializer := awsdeserializer.NewS3BucketAnalyticDeserializer()
			s := &S3BucketAnalyticSupplier{
				provider,
				deserializer,
				factory,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			if (err != nil) != tt.wantErr {
				t.Errorf("Resources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			test.CtyTestDiff(got, tt.dirName, provider, deserializer, shouldUpdate, t)
		})
	}
}
