package remote

import (
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/aws"
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/terraform"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration/resource"
	resourceaws "github.com/snyk/driftctl/enumeration/resource/aws"
	"github.com/snyk/driftctl/mocks"

	"github.com/snyk/driftctl/test/goldenfile"
	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSQSQueue(t *testing.T) {
	cases := []struct {
		test           string
		dirName        string
		mocks          func(*repository.MockSQSRepository, *mocks.AlerterInterface)
		assertExpected func(*testing.T, []*resource.Resource)
		wantErr        error
	}{
		{
			test:    "no sqs queues",
			dirName: "aws_sqs_queue_empty",
			mocks: func(client *repository.MockSQSRepository, alerter *mocks.AlerterInterface) {
				client.On("ListAllQueues").Return([]*string{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: nil,
		},
		{
			test:    "multiple sqs queues",
			dirName: "aws_sqs_queue_multiple",
			mocks: func(client *repository.MockSQSRepository, alerter *mocks.AlerterInterface) {
				client.On("ListAllQueues").Return([]*string{
					awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/bar.fifo"),
					awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/foo"),
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, "https://sqs.eu-west-3.amazonaws.com/047081014315/bar.fifo", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsSqsQueueResourceType, got[0].ResourceType())

				assert.Equal(t, "https://sqs.eu-west-3.amazonaws.com/047081014315/foo", got[1].ResourceId())
				assert.Equal(t, resourceaws.AwsSqsQueueResourceType, got[1].ResourceType())
			},
			wantErr: nil,
		},
		{
			test:    "cannot list sqs queues",
			dirName: "aws_sqs_queue_empty",
			mocks: func(client *repository.MockSQSRepository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				client.On("ListAllQueues").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsSqsQueueResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsSqsQueueResourceType, resourceaws.AwsSqsQueueResourceType), alerts.EnumerationPhase)).Return()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: nil,
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

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

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			fakeRepo.AssertExpectations(tt)
			alerter.AssertExpectations(tt)
		})
	}
}

func TestSQSQueuePolicy(t *testing.T) {
	cases := []struct {
		test           string
		dirName        string
		mocks          func(*repository.MockSQSRepository, *mocks.AlerterInterface)
		assertExpected func(*testing.T, []*resource.Resource)
		wantErr        error
	}{
		{
			// sqs queue with no policy case is not possible
			// as a default SQSDefaultPolicy (e.g. policy="") will always be present in each queue
			test:    "no sqs queue policies",
			dirName: "aws_sqs_queue_policy_empty",
			mocks: func(client *repository.MockSQSRepository, alerter *mocks.AlerterInterface) {
				client.On("ListAllQueues").Return([]*string{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: nil,
		},
		{
			test:    "multiple sqs queue policies (default or not)",
			dirName: "aws_sqs_queue_policy_multiple",
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
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 3)

				assert.Equal(t, "https://sqs.eu-west-3.amazonaws.com/047081014315/bar.fifo", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsSqsQueuePolicyResourceType, got[0].ResourceType())

				assert.Equal(t, "https://sqs.eu-west-3.amazonaws.com/047081014315/foo", got[1].ResourceId())
				assert.Equal(t, resourceaws.AwsSqsQueuePolicyResourceType, got[1].ResourceType())

				assert.Equal(t, "https://sqs.eu-west-3.amazonaws.com/047081014315/baz", got[2].ResourceId())
				assert.Equal(t, resourceaws.AwsSqsQueuePolicyResourceType, got[2].ResourceType())
			},
			wantErr: nil,
		},
		{
			test:    "multiple sqs queue policies (with nil attributes)",
			dirName: "aws_sqs_queue_policy_multiple",
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
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 3)

				assert.Equal(t, "https://sqs.eu-west-3.amazonaws.com/047081014315/bar.fifo", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsSqsQueuePolicyResourceType, got[0].ResourceType())

				assert.Equal(t, "https://sqs.eu-west-3.amazonaws.com/047081014315/foo", got[1].ResourceId())
				assert.Equal(t, resourceaws.AwsSqsQueuePolicyResourceType, got[1].ResourceType())

				assert.Equal(t, "https://sqs.eu-west-3.amazonaws.com/047081014315/baz", got[2].ResourceId())
				assert.Equal(t, resourceaws.AwsSqsQueuePolicyResourceType, got[2].ResourceType())
			},
			wantErr: nil,
		},
		{
			test:    "cannot list sqs queues, thus sqs queue policies",
			dirName: "aws_sqs_queue_policy_empty",
			mocks: func(client *repository.MockSQSRepository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				client.On("ListAllQueues").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsSqsQueuePolicyResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsSqsQueuePolicyResourceType, resourceaws.AwsSqsQueueResourceType), alerts.EnumerationPhase)).Return()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: nil,
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

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

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			fakeRepo.AssertExpectations(tt)
			alerter.AssertExpectations(tt)
		})
	}
}
