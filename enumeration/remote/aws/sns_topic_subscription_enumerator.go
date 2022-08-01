package aws

import (
	"fmt"

	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type wrongArnTopicAlert struct {
	arn      string
	endpoint *string
}

func NewWrongArnTopicAlert(arn string, endpoint *string) *wrongArnTopicAlert {
	return &wrongArnTopicAlert{arn: arn, endpoint: endpoint}
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

type SNSTopicSubscriptionEnumerator struct {
	repository repository.SNSRepository
	factory    resource.ResourceFactory
	alerter    alerter.AlerterInterface
}

func NewSNSTopicSubscriptionEnumerator(
	repo repository.SNSRepository,
	factory resource.ResourceFactory,
	alerter alerter.AlerterInterface,
) *SNSTopicSubscriptionEnumerator {
	return &SNSTopicSubscriptionEnumerator{
		repo,
		factory,
		alerter,
	}
}

func (e *SNSTopicSubscriptionEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsSnsTopicSubscriptionResourceType
}

func (e *SNSTopicSubscriptionEnumerator) Enumerate() ([]*resource.Resource, error) {
	allSubscriptions, err := e.repository.ListAllSubscriptions()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(allSubscriptions))

	for _, subscription := range allSubscriptions {
		if subscription.SubscriptionArn == nil || !arn.IsARN(*subscription.SubscriptionArn) {
			e.alerter.SendAlert(
				fmt.Sprintf("%s.%s", e.SupportedType(), *subscription.SubscriptionArn),
				NewWrongArnTopicAlert(*subscription.SubscriptionArn, subscription.Endpoint),
			)
			continue
		}
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*subscription.SubscriptionArn,
				map[string]interface{}{
					"SubscriptionId": *subscription.SubscriptionArn,
				},
			),
		)
	}

	return results, nil
}
