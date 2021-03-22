package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type EKSClusterDeserializer struct {
}

func NewEKSClusterDeserializer() *EKSClusterDeserializer {
	return &EKSClusterDeserializer{}
}

func (s *EKSClusterDeserializer) HandledType() resource.ResourceType {
	return aws.AwsLambdaEventSourceMappingResourceType
}

func (s EKSClusterDeserializer) Deserialize(values []cty.Value) ([]resource.Resource, error) {
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

func decodeEKSCluster(value cty.Value) (resource.Resource, error) {
	var eventSourceMapping aws.AwsLambdaEventSourceMapping
	err := gocty.FromCtyValue(value, &eventSourceMapping)
	return &eventSourceMapping, err
}
