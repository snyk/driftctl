package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/helpers"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/zclconf/go-cty/cty"
)

type SNSTopicPolicyDeserializer struct {
	deserializer
}

func NewSNSTopicPolicyDeserializer() *SNSTopicPolicyDeserializer {
	return &SNSTopicPolicyDeserializer{
		deserializer{
			normalize: func(res resource.Resource) error {
				r := res.(*aws.AwsSnsTopicPolicy)
				if r.Policy != nil {
					jsonString, err := helpers.NormalizeJsonString(*r.Policy)
					if err != nil {
						return err
					}
					r.Policy = &jsonString
				}
				return nil
			},
		},
	}
}

func (s *SNSTopicPolicyDeserializer) HandledType() resource.ResourceType {
	return aws.AwsSnsTopicPolicyResourceType
}

func (s SNSTopicPolicyDeserializer) Deserialize(topicList []cty.Value) ([]resource.Resource, error) {
	return s.deserialize(topicList, &aws.AwsSnsTopicPolicy{})
}
