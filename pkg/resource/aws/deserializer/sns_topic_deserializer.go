package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type SNSTopicDeserializer struct {
}

func NewSNSTopicDeserializer() *SNSTopicDeserializer {
	return &SNSTopicDeserializer{}
}

func (s *SNSTopicDeserializer) HandledType() resource.ResourceType {
	return aws.AwsSnsTopicResourceType
}

func (s SNSTopicDeserializer) Deserialize(topicList []cty.Value) ([]resource.Resource, error) {
	topics := make([]resource.Resource, 0)

	for _, value := range topicList {
		topic, err := decodeSNSTopic(value)
		if err != nil {
			return nil, err
		}
		topics = append(topics, topic)
	}
	return topics, nil
}

func decodeSNSTopic(value cty.Value) (resource.Resource, error) {
	var topic aws.AwsSnsTopic
	err := gocty.FromCtyValue(value, &topic)
	return &topic, err
}
