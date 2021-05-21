package aws

import (
	"fmt"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"

	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type pendingTopicAlert struct {
	endpoint *string
}

func (p *pendingTopicAlert) Message() string {
	return fmt.Sprintf("%s with pending confirmation status for endpoint \"%s\" will be ignored",
		aws.AwsSnsTopicSubscriptionResourceType,
		awssdk.StringValue(p.endpoint))
}

func (p *pendingTopicAlert) ShouldIgnoreResource() bool {
	return false
}

type SNSTopicSubscriptionSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	client       repository.SNSRepository
	runner       *terraform.ParallelResourceReader
	alerter      alerter.AlerterInterface
}

func NewSNSTopicSubscriptionSupplier(provider *AWSTerraformProvider, a alerter.AlerterInterface, deserializer *resource.Deserializer) *SNSTopicSubscriptionSupplier {
	return &SNSTopicSubscriptionSupplier{
		provider,
		deserializer,
		repository.NewSNSClient(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
		a,
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
			return s.readTopicSubscription(subscription, s.alerter)
		})
	}

	retrieve, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(aws.AwsSnsTopicSubscriptionResourceType, retrieve)
}

func (s *SNSTopicSubscriptionSupplier) readTopicSubscription(subscription *sns.Subscription, alertr alerter.AlerterInterface) (cty.Value, error) {
	if subscription.SubscriptionArn != nil && *subscription.SubscriptionArn == "PendingConfirmation" {
		alertr.SendAlert(
			fmt.Sprintf("%s.%s", aws.AwsSnsTopicSubscriptionResourceType, *subscription.SubscriptionArn),
			&pendingTopicAlert{subscription.Endpoint},
		)
		return cty.NilVal, nil
	}

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
