package remote

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/mocks"
	"github.com/snyk/driftctl/pkg/filter"
	"github.com/snyk/driftctl/pkg/remote/alerts"
	"github.com/snyk/driftctl/pkg/remote/aws"
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	"github.com/snyk/driftctl/pkg/remote/cache"
	"github.com/snyk/driftctl/pkg/remote/common"
	remoteerr "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	resourceaws "github.com/snyk/driftctl/pkg/resource/aws"
	"github.com/snyk/driftctl/pkg/terraform"
	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/goldenfile"
	testresource "github.com/snyk/driftctl/test/resource"
	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSQSQueue(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockSQSRepository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no sqs queues",
			dirName: "sqs_queue_empty",
			mocks: func(client *repository.MockSQSRepository, alerter *mocks.AlerterInterface) {
				client.On("ListAllQueues").Return([]*string{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "multiple sqs queues",
			dirName: "sqs_queue_multiple",
			mocks: func(client *repository.MockSQSRepository, alerter *mocks.AlerterInterface) {
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
			mocks: func(client *repository.MockSQSRepository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				client.On("ListAllQueues").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsSqsQueueResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsSqsQueueResourceType, resourceaws.AwsSqsQueueResourceType), alerts.EnumerationPhase)).Return()
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
			fakeRepo := &repository.MockSQSRepository{}
			c.mocks(fakeRepo, alerter)
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
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsSqsQueueResourceType, aws.NewSQSQueueDetailsFetcher(provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsSqsQueueResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			fakeRepo.AssertExpectations(tt)
			alerter.AssertExpectations(tt)
		})
	}
}

func TestSQSQueuePolicy(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockSQSRepository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			// sqs queue with no policy case is not possible
			// as a default SQSDefaultPolicy (e.g. policy="") will always be present in each queue
			test:    "no sqs queue policies",
			dirName: "sqs_queue_policy_empty",
			mocks: func(client *repository.MockSQSRepository, alerter *mocks.AlerterInterface) {
				client.On("ListAllQueues").Return([]*string{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "multiple sqs queue policies (default or not)",
			dirName: "sqs_queue_policy_multiple",
			mocks: func(client *repository.MockSQSRepository, alerter *mocks.AlerterInterface) {
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
			test:    "multiple sqs queue policies (with nil attributes)",
			dirName: "sqs_queue_policy_multiple",
			mocks: func(client *repository.MockSQSRepository, alerter *mocks.AlerterInterface) {
				client.On("ListAllQueues").Return([]*string{
					awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/bar.fifo"),
					awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/foo"),
					awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/baz"),
				}, nil)

				client.On("GetQueueAttributes", mock.Anything).Return(
					&sqs.GetQueueAttributesOutput{},
					nil,
				)
			},
			wantErr: nil,
		},
		{
			test:    "cannot list sqs queues, thus sqs queue policies",
			dirName: "sqs_queue_policy_empty",
			mocks: func(client *repository.MockSQSRepository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				client.On("ListAllQueues").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsSqsQueuePolicyResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsSqsQueuePolicyResourceType, resourceaws.AwsSqsQueueResourceType), alerts.EnumerationPhase)).Return()
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
			fakeRepo := &repository.MockSQSRepository{}
			c.mocks(fakeRepo, alerter)
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

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsSqsQueuePolicyResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			fakeRepo.AssertExpectations(tt)
			alerter.AssertExpectations(tt)
		})
	}
}
