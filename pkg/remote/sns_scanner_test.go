package remote

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/remote/aws"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	terraform2 "github.com/cloudskiff/driftctl/test/terraform"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	"github.com/stretchr/testify/assert"
)

func TestScanSNSTopic(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(client *repository.MockSNSRepository)
		err     error
	}{
		{
			test:    "no SNS Topic",
			dirName: "sns_topic_empty",
			mocks: func(client *repository.MockSNSRepository) {
				client.On("ListAllTopics").Return([]*sns.Topic{}, nil)
			},
			err: nil,
		},
		{
			test:    "Multiple SNSTopic",
			dirName: "sns_topic_multiple",
			mocks: func(client *repository.MockSNSRepository) {
				client.On("ListAllTopics").Return([]*sns.Topic{
					{TopicArn: awssdk.String("arn:aws:sns:eu-west-3:526954929923:user-updates-topic")},
					{TopicArn: awssdk.String("arn:aws:sns:eu-west-3:526954929923:user-updates-topic2")},
					{TopicArn: awssdk.String("arn:aws:sns:eu-west-3:526954929923:user-updates-topic3")},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list SNSTopic",
			dirName: "sns_topic_empty",
			mocks: func(client *repository.MockSNSRepository) {
				client.On("ListAllTopics").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			session := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			alerter := &mocks.AlerterInterface{}
			alerter.On("SendAlert", mock.Anything, mock.Anything).Maybe().Return()

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			fakeRepo := &repository.MockSNSRepository{}
			c.mocks(fakeRepo)
			var repo repository.SNSRepository = fakeRepo
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
				repo = repository.NewSNSRepository(session, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewSNSTopicEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsSnsTopicResourceType, aws.NewSNSTopicDetailsFetcher(provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsSnsTopicResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestSNSTopicPolicyScan(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(client *repository.MockSNSRepository)
		err     error
	}{
		{
			test:    "no SNS Topic policy",
			dirName: "sns_topic_policy_empty",
			mocks: func(client *repository.MockSNSRepository) {
				client.On("ListAllTopics").Return([]*sns.Topic{}, nil)
			},
			err: nil,
		},
		{
			test:    "Multiple SNSTopicPolicy",
			dirName: "sns_topic_policy_multiple",
			mocks: func(client *repository.MockSNSRepository) {
				client.On("ListAllTopics").Return([]*sns.Topic{
					{TopicArn: awssdk.String("arn:aws:sns:us-east-1:526954929923:my-topic-with-policy")},
					{TopicArn: awssdk.String("arn:aws:sns:us-east-1:526954929923:my-topic-with-policy2")},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list SNSTopic",
			dirName: "sns_topic_policy_topic_list",
			mocks: func(client *repository.MockSNSRepository) {
				client.On("ListAllTopics").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			session := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			alerter := &mocks.AlerterInterface{}
			alerter.On("SendAlert", mock.Anything, mock.Anything).Maybe().Return()

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			fakeRepo := &repository.MockSNSRepository{}
			c.mocks(fakeRepo)
			var repo repository.SNSRepository = fakeRepo
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
				repo = repository.NewSNSRepository(session, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewSNSTopicPolicyEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsSnsTopicPolicyResourceType, aws.NewSNSTopicPolicyDetailsFetcher(provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsSnsTopicPolicyResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}
