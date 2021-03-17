package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type SNSTopicSubscriptionDeserializer struct {
}

func NewSNSTopicSubscriptionDeserializer() *SNSTopicSubscriptionDeserializer {
	return &SNSTopicSubscriptionDeserializer{}
}

func (s *SNSTopicSubscriptionDeserializer) HandledType() resource.ResourceType {
	return aws.AwsSnsTopicSubscriptionResourceType
}

func (s SNSTopicSubscriptionDeserializer) Deserialize(subscriptionsList []cty.Value) ([]resource.Resource, error) {
	subscriptions := make([]resource.Resource, 0)

	for _, value := range subscriptionsList {

		value := value
		subscription, err := decodeSNSTopicSubscription(value)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, subscription)
	}
	return subscriptions, nil
}

func decodeSNSTopicSubscription(value cty.Value) (resource.Resource, error) {
	var subscription aws.AwsSnsTopicSubscription
	err := gocty.FromCtyValue(value, &subscription)
	return &subscription, err
}
