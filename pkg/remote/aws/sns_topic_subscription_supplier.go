package aws

import (
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"

	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type SNSTopicSubscriptionSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       repository.SNSRepository
	runner       *terraform.ParallelResourceReader
}

func NewSNSTopicSubscriptionSupplier(provider *AWSTerraformProvider) *SNSTopicSubscriptionSupplier {
	return &SNSTopicSubscriptionSupplier{
		provider,
		awsdeserializer.NewSNSTopicSubscriptionDeserializer(),
		repository.NewSNSClient(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *SNSTopicSubscriptionSupplier) Resources() ([]resource.Resource, error) {
	subscriptions, err := s.client.ListAllSubscriptions()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsSnsTopicSubscriptionResourceType)
	}
	for _, subscription := range subscriptions {
		subscription := subscription
		s.runner.Run(func() (cty.Value, error) {
			return s.readTopicSubscription(subscription)
		})
	}

	retrieve, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(retrieve)
}

func (s *SNSTopicSubscriptionSupplier) readTopicSubscription(subscription *sns.Subscription) (cty.Value, error) {
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		ID: *subscription.SubscriptionArn,
		Ty: aws.AwsSnsTopicSubscriptionResourceType,
		Attributes: map[string]string{
			"SubscriptionId": *subscription.SubscriptionArn,
		},
	})
	if err != nil {
		logrus.Error(err)
		return cty.NilVal, err
	}
	return *val, nil
}
