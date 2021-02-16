package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/zclconf/go-cty/cty"
)

type DynamoDBTableDeserializer struct {
	deserializer
}

func NewDynamoDBTableDeserializer() *DynamoDBTableDeserializer {
	return &DynamoDBTableDeserializer{}
}

func (s *DynamoDBTableDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsDynamodbTableResourceType
}

func (s DynamoDBTableDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	return s.deserialize(rawList, &resourceaws.AwsDynamodbTable{})
}
