package aws

import (
	"fmt"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
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

type wrongArnTopicAlert struct {
	arn      string
	endpoint *string
}

func (p *wrongArnTopicAlert) Message() string {
	return fmt.Sprintf("%s with incorrect subscription arn (%s) for endpoint \"%s\" will be ignored",
		aws.AwsSnsTopicSubscriptionResourceType,
		p.arn,
		awssdk.StringValue(p.endpoint))
}

func (p *wrongArnTopicAlert) ShouldIgnoreResource() bool {
	return false
}

type SNSTopicSubscriptionSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	client       repository.SNSRepository
	runner       *terraform.ParallelResourceReader
	alerter      alerter.AlerterInterface
}

func NewSNSTopicSubscriptionSupplier(provider *AWSTerraformProvider, a alerter.AlerterInterface, deserializer *resource.Deserializer, client repository.SNSRepository) *SNSTopicSubscriptionSupplier {
	return &SNSTopicSubscriptionSupplier{
		provider,
		deserializer,
		client,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
		a,
	}
}

func (s *SNSTopicSubscriptionSupplier) SuppliedType() resource.ResourceType {
	return aws.AwsSnsTopicSubscriptionResourceType
}

func (s *SNSTopicSubscriptionSupplier) Resources() ([]resource.Resource, error) {
	subscriptions, err := s.client.ListAllSubscriptions()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
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

	return s.deserializer.Deserialize(s.SuppliedType(), retrieve)
}

func (s *SNSTopicSubscriptionSupplier) readTopicSubscription(subscription *sns.Subscription, alertr alerter.AlerterInterface) (cty.Value, error) {
	if subscription.SubscriptionArn != nil && !arn.IsARN(*subscription.SubscriptionArn) {
		alertr.SendAlert(
			fmt.Sprintf("%s.%s", s.SuppliedType(), *subscription.SubscriptionArn),
			&wrongArnTopicAlert{*subscription.SubscriptionArn, subscription.Endpoint},
		)
		return cty.NilVal, nil
	}

	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		ID: *subscription.SubscriptionArn,
		Ty: s.SuppliedType(),
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
