package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type LambdaEventSourceMappingTopicDeserializer struct {
}

func NewLambdaEventSourceMappingDeserializer() *LambdaEventSourceMappingTopicDeserializer {
	return &LambdaEventSourceMappingTopicDeserializer{}
}

func (s *LambdaEventSourceMappingTopicDeserializer) HandledType() resource.ResourceType {
	return aws.AwsLambdaEventSourceMappingResourceType
}

func (s LambdaEventSourceMappingTopicDeserializer) Deserialize(values []cty.Value) ([]resource.Resource, error) {
	decoded := make([]resource.Resource, 0)

	for _, value := range values {
		eventSourceMapping, err := decodeLambdaEventSourceMapping(value)
		if err != nil {
			return nil, err
		}
		decoded = append(decoded, eventSourceMapping)
	}
	return decoded, nil
}

func decodeLambdaEventSourceMapping(value cty.Value) (resource.Resource, error) {
	var eventSourceMapping aws.AwsLambdaEventSourceMapping
	err := gocty.FromCtyValue(value, &eventSourceMapping)
	return &eventSourceMapping, err
}
