package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	testresource "github.com/cloudskiff/driftctl/test/resource"

	awssdk "github.com/aws/aws-sdk-go/aws"

	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/cloudskiff/driftctl/pkg/resource"

	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	mocks2 "github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSqsQueuePolicySupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *mocks.SQSRepository)
		err     error
	}{
		{
			// sqs queue with no policy case is not possible
			// as a default SQSDefaultPolicy (e.g. policy="") will always be present in each queue
			test:    "no sqs queue policies",
			dirName: "sqs_queue_policy_empty",
			mocks: func(client *mocks.SQSRepository) {
				client.On("ListAllQueues").Return([]*string{}, nil)
			},
			err: nil,
		},
		{
			test:    "multiple sqs queue policies (default or not)",
			dirName: "sqs_queue_policy_multiple",
			mocks: func(client *mocks.SQSRepository) {
				client.On("ListAllQueues").Return([]*string{
					awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/bar.fifo"),
					awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/foo"),
					awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/baz"),
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list sqs queues, thus sqs queue policies",
			dirName: "sqs_queue_policy_empty",
			mocks: func(client *mocks.SQSRepository) {
				client.On("ListAllQueues").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationErrorWithType(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsSqsQueuePolicyResourceType, resourceaws.AwsSqsQueueResourceType),
		},
	}
	for _, c := range cases {
		shouldUpdate := c.dirName == *goldenfile.Update
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
			supplierLibrary.AddSupplier(NewSqsQueuePolicySupplier(provider, deserializer))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeSQS := mocks.SQSRepository{}
			c.mocks(&fakeSQS)
			provider := mocks2.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			s := &SqsQueuePolicySupplier{
				provider,
				deserializer,
				&fakeSQS,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)

			mock.AssertExpectationsForObjects(tt)
			test.CtyTestDiff(got, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}
