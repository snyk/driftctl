package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/zclconf/go-cty/cty"
)

type SNSTopicDeserializer struct {
	deserializer
}

func NewSNSTopicDeserializer() *SNSTopicDeserializer {
	return &SNSTopicDeserializer{}
}

func (s *SNSTopicDeserializer) HandledType() resource.ResourceType {
	return aws.AwsSnsTopicResourceType
}

func (s SNSTopicDeserializer) Deserialize(topicList []cty.Value) ([]resource.Resource, error) {
	return s.deserialize(topicList, &aws.AwsSnsTopic{})
}
