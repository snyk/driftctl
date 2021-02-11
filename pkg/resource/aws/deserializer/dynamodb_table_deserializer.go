package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type DynamoDBTableDeserializer struct {
}

func NewDynamoDBTableDeserializer() *DynamoDBTableDeserializer {
	return &DynamoDBTableDeserializer{}
}

func (s *DynamoDBTableDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsDynamodbTableResourceType
}

func (s DynamoDBTableDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range rawList {
		rawResource := rawResource
		resource, err := decodeDynamoDBTable(&rawResource)
		if err != nil {
			logrus.Warnf("Error when deserializing resource %+v : %+v", rawResource, err)
			return nil, err
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func decodeDynamoDBTable(raw *cty.Value) (*resourceaws.AwsDynamodbTable, error) {
	var decoded resourceaws.AwsDynamodbTable
	if err := gocty.FromCtyValue(*raw, &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}
