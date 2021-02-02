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

func TestSNSTopicSupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *mocks.SNSClient)
		err     error
	}{
		{
			test:    "no SNS Topic",
			dirName: "sns_topic_empty",
			mocks: func(client *mocks.SNSClient) {
				client.On("ListAllTopics").Return([]*sns.Topic{}, nil)
			},
			err: nil,
		},
		{
			test:    "Multiple SNSTopic",
			dirName: "sns_topic_multiple",
			mocks: func(client *mocks.SNSClient) {
				client.On("ListAllTopics").Return([]*sns.Topic{
					{TopicArn: aws.String("arn:aws:sns:eu-west-3:526954929923:user-updates-topic")},
					{TopicArn: aws.String("arn:aws:sns:eu-west-3:526954929923:user-updates-topic2")},
					{TopicArn: aws.String("arn:aws:sns:eu-west-3:526954929923:user-updates-topic3")},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list SNSTopic",
			dirName: "sns_topic_empty",
			mocks: func(client *mocks.SNSClient) {
				client.On("ListAllTopics").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsSnsTopicResourceType),
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
			supplierLibrary.AddSupplier(NewSNSTopicSupplier(provider))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeClient := mocks.SNSClient{}
			c.mocks(&fakeClient)
			provider := mocks2.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			SNSTopicDeserializer := awsdeserializer.NewSNSTopicDeserializer()
			s := &SNSTopicSupplier{
				provider,
				SNSTopicDeserializer,
				&fakeClient,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)

			mock.AssertExpectationsForObjects(tt)
			test.CtyTestDiff(got, c.dirName, provider, SNSTopicDeserializer, shouldUpdate, tt)
		})
	}
}
