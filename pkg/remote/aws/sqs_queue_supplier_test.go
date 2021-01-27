package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/service/sqs"

	awssdk "github.com/aws/aws-sdk-go/aws"

	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	mocks2 "github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSqsQueueSupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *mocks.FakeSQS)
		err     error
	}{
		{
			test:    "no sqs queues",
			dirName: "sqs_queue_empty",
			mocks: func(client *mocks.FakeSQS) {
				client.On("ListQueuesPages",
					&sqs.ListQueuesInput{},
					mock.MatchedBy(func(callback func(res *sqs.ListQueuesOutput, lastPage bool) bool) bool {
						callback(&sqs.ListQueuesOutput{}, true)
						return true
					})).Return(nil)
			},
			err: nil,
		},
		{
			test:    "multiple sqs queues",
			dirName: "sqs_queue_multiple",
			mocks: func(client *mocks.FakeSQS) {
				client.On("ListQueuesPages",
					&sqs.ListQueuesInput{},
					mock.MatchedBy(func(callback func(res *sqs.ListQueuesOutput, lastPage bool) bool) bool {
						callback(&sqs.ListQueuesOutput{
							QueueUrls: []*string{
								awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/bar.fifo"),
								awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/foo"),
							},
						}, true)
						return true
					})).Return(nil)
			},
			err: nil,
		},
		{
			test:    "cannot list sqs queues",
			dirName: "sqs_queue_empty",
			mocks: func(client *mocks.FakeSQS) {
				client.On(
					"ListQueuesPages",
					&sqs.ListQueuesInput{},
					mock.Anything,
				).Once().Return(awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsSqsQueueResourceType),
		},
	}
	for _, c := range cases {
		shouldUpdate := c.dirName == *goldenfile.Update
		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		if shouldUpdate {
			provider, err := NewTerraFormProvider()
			if err != nil {
				t.Fatal(err)
			}

			providerLibrary.AddProvider(terraform.AWS, provider)
			supplierLibrary.AddSupplier(NewSqsQueueSupplier(provider))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeSQS := mocks.FakeSQS{}
			c.mocks(&fakeSQS)
			provider := mocks2.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			sqsQueueDeserializer := awsdeserializer.NewSqsQueueDeserializer()
			s := &SqsQueueSupplier{
				provider,
				sqsQueueDeserializer,
				&fakeSQS,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)

			mock.AssertExpectationsForObjects(tt)
			deserializers := []deserializer.CTYDeserializer{sqsQueueDeserializer}
			test.CtyTestDiffMixed(got, c.dirName, provider, deserializers, shouldUpdate, tt)
		})
	}
}
