package remote

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/remote/aws"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	terraform2 "github.com/cloudskiff/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSQSQueue(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *repository.MockSQSRepository)
		wantErr error
	}{
		{
			test:    "no sqs queues",
			dirName: "sqs_queue_empty",
			mocks: func(client *repository.MockSQSRepository) {
				client.On("ListAllQueues").Return([]*string{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "multiple sqs queues",
			dirName: "sqs_queue_multiple",
			mocks: func(client *repository.MockSQSRepository) {
				client.On("ListAllQueues").Return([]*string{
					awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/bar.fifo"),
					awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/foo"),
				}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "cannot list sqs queues",
			dirName: "sqs_queue_empty",
			mocks: func(client *repository.MockSQSRepository) {
				client.On("ListAllQueues").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			alerter.On("SendAlert", mock.Anything, mock.Anything).Maybe().Return()
			fakeRepo := &repository.MockSQSRepository{}
			c.mocks(fakeRepo)
			var repo repository.SQSRepository = fakeRepo
			providerVersion := "3.19.0"
			realProvider, err := terraform2.InitTestAwsProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				repo = repository.NewSQSRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewSQSQueueEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsSqsQueueResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsSqsQueueResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsSqsQueueResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestSQSQueuePolicy(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *repository.MockSQSRepository)
		wantErr error
	}{
		{
			// sqs queue with no policy case is not possible
			// as a default SQSDefaultPolicy (e.g. policy="") will always be present in each queue
			test:    "no sqs queue policies",
			dirName: "sqs_queue_policy_empty",
			mocks: func(client *repository.MockSQSRepository) {
				client.On("ListAllQueues").Return([]*string{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "multiple sqs queue policies (default or not)",
			dirName: "sqs_queue_policy_multiple",
			mocks: func(client *repository.MockSQSRepository) {
				client.On("ListAllQueues").Return([]*string{
					awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/bar.fifo"),
					awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/foo"),
					awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/baz"),
				}, nil)

				client.On("GetQueueAttributes", mock.Anything).Return(
					&sqs.GetQueueAttributesOutput{
						Attributes: map[string]*string{
							sqs.QueueAttributeNamePolicy: awssdk.String(""),
						},
					},
					nil,
				)
			},
			wantErr: nil,
		},
		{
			test:    "cannot list sqs queues, thus sqs queue policies",
			dirName: "sqs_queue_policy_empty",
			mocks: func(client *repository.MockSQSRepository) {
				client.On("ListAllQueues").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			alerter.On("SendAlert", mock.Anything, mock.Anything).Maybe().Return()
			fakeRepo := &repository.MockSQSRepository{}
			c.mocks(fakeRepo)
			var repo repository.SQSRepository = fakeRepo
			providerVersion := "3.19.0"
			realProvider, err := terraform2.InitTestAwsProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				repo = repository.NewSQSRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewSQSQueuePolicyEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsSqsQueuePolicyResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsSqsQueuePolicyResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsSqsQueuePolicyResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}
