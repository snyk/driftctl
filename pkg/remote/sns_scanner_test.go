package remote

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/filter"
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

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
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

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsSnsTopicPolicyResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestSNSTopicSubscriptionScan(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(client *repository.MockSNSRepository)
		alerts  alerter.Alerts
		err     error
	}{
		{
			test:    "no SNS Topic Subscription",
			dirName: "sns_topic_subscription_empty",
			mocks: func(client *repository.MockSNSRepository) {
				client.On("ListAllSubscriptions").Return([]*sns.Subscription{}, nil)
			},
			alerts: map[string][]alerter.Alert{},
			err:    nil,
		},
		{
			test:    "Multiple SNSTopic Subscription",
			dirName: "sns_topic_subscription_multiple",
			mocks: func(client *repository.MockSNSRepository) {
				client.On("ListAllSubscriptions").Return([]*sns.Subscription{
					{SubscriptionArn: awssdk.String("arn:aws:sns:us-east-1:526954929923:user-updates-topic2:c0f794c5-a009-4db4-9147-4c55959787fa")},
					{SubscriptionArn: awssdk.String("arn:aws:sns:us-east-1:526954929923:user-updates-topic:b6e66147-2b31-4486-8d4b-2a2272264c8e")},
				}, nil)
			},
			alerts: map[string][]alerter.Alert{},
			err:    nil,
		},
		{
			test:    "Multiple SNSTopic Subscription with one pending and one incorrect",
			dirName: "sns_topic_subscription_multiple",
			mocks: func(client *repository.MockSNSRepository) {
				client.On("ListAllSubscriptions").Return([]*sns.Subscription{
					{SubscriptionArn: awssdk.String("PendingConfirmation"), Endpoint: awssdk.String("TEST")},
					{SubscriptionArn: awssdk.String("Incorrect"), Endpoint: awssdk.String("INCORRECT")},
					{SubscriptionArn: awssdk.String("arn:aws:sns:us-east-1:526954929923:user-updates-topic2:c0f794c5-a009-4db4-9147-4c55959787fa")},
					{SubscriptionArn: awssdk.String("arn:aws:sns:us-east-1:526954929923:user-updates-topic:b6e66147-2b31-4486-8d4b-2a2272264c8e")},
				}, nil)
			},
			alerts: map[string][]alerter.Alert{
				"aws_sns_topic_subscription.PendingConfirmation": []alerter.Alert{
					aws.NewWrongArnTopicAlert("PendingConfirmation", awssdk.String("TEST")),
				},
				"aws_sns_topic_subscription.Incorrect": []alerter.Alert{
					aws.NewWrongArnTopicAlert("Incorrect", awssdk.String("INCORRECT")),
				},
			},
			err: nil,
		},
		{
			test:    "cannot list SNSTopic subscription",
			dirName: "sns_topic_subscription_list",
			mocks: func(client *repository.MockSNSRepository) {
				client.On("ListAllSubscriptions").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			alerts: map[string][]alerter.Alert{
				resourceaws.AwsSnsTopicSubscriptionResourceType: {
					NewEnumerationAccessDeniedAlert("aws+tf", resourceaws.AwsSnsTopicSubscriptionResourceType, resourceaws.AwsSnsTopicSubscriptionResourceType),
				},
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

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := alerter.NewAlerter()
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

			remoteLibrary.AddEnumerator(aws.NewSNSTopicSubscriptionEnumerator(repo, factory, alerter))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsSnsTopicSubscriptionResourceType, aws.NewSNSTopicSubscriptionDetailsFetcher(provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.err)
			if err != nil {
				return
			}
			assert.Equal(tt, c.alerts, alerter.Retrieve())
			test.TestAgainstGoldenFile(got, resourceaws.AwsSnsTopicSubscriptionResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}
