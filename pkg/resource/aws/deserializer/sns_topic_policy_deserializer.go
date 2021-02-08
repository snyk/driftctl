package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type SNSTopicPolicyDeserializer struct {
}

func NewSNSTopicPolicyDeserializer() *SNSTopicPolicyDeserializer {
	return &SNSTopicPolicyDeserializer{}
}

func (s *SNSTopicPolicyDeserializer) HandledType() resource.ResourceType {
	return aws.AwsSnsTopicPolicyResourceType
}

func (s SNSTopicPolicyDeserializer) Deserialize(topicList []cty.Value) ([]resource.Resource, error) {
	policies := make([]resource.Resource, 0)

	for _, value := range topicList {
		value := value
		policy, err := decodeSNSTopicPolicy(value)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}
	return policies, nil
}

func decodeSNSTopicPolicy(value cty.Value) (resource.Resource, error) {
	var topicPolicy aws.AwsSnsTopicPolicy
	err := gocty.FromCtyValue(value, &topicPolicy)
	return &topicPolicy, err
}
