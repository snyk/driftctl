package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/sns"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/cloudskiff/driftctl/pkg/parallel"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/cloudskiff/driftctl/test/goldenfile"
	mocks2 "github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/mocks"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
)

func TestSNSTopicSubscriptionSupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *mocks.SNSRepository)
		err     error
	}{
		{
			test:    "no SNS Topic Subscription",
			dirName: "sns_topic_subscription_empty",
			mocks: func(client *mocks.SNSRepository) {
				client.On("ListAllSubscriptions").Return([]*sns.Subscription{}, nil)
			},
			err: nil,
		},
		{
			test:    "Multiple SNSTopic Subscription",
			dirName: "sns_topic_subscription_multiple",
			mocks: func(client *mocks.SNSRepository) {
				client.On("ListAllSubscriptions").Return([]*sns.Subscription{
					{SubscriptionArn: aws.String("arn:aws:sns:us-east-1:526954929923:user-updates-topic2:c0f794c5-a009-4db4-9147-4c55959787fa")},
					{SubscriptionArn: aws.String("arn:aws:sns:us-east-1:526954929923:user-updates-topic:b6e66147-2b31-4486-8d4b-2a2272264c8e")},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list SNSTopic subscription",
			dirName: "sns_topic_subscription_list",
			mocks: func(client *mocks.SNSRepository) {
				client.On("ListAllSubscriptions").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsSnsTopicSubscriptionResourceType),
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
			supplierLibrary.AddSupplier(NewSNSTopicSubscriptionSupplier(provider))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeClient := mocks.SNSRepository{}
			c.mocks(&fakeClient)
			provider := mocks2.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			topicSubscriptionDeserializer := awsdeserializer.NewSNSTopicSubscriptionDeserializer()
			s := &SNSTopicSubscriptionSupplier{
				provider,
				topicSubscriptionDeserializer,
				&fakeClient,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)

			mock.AssertExpectationsForObjects(tt)
			test.CtyTestDiff(got, c.dirName, provider, topicSubscriptionDeserializer, shouldUpdate, tt)
		})
	}
}
