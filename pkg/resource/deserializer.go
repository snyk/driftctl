package resource

import (
	"encoding/json"

	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

type Deserializer struct {
	factory ResourceFactory
}

func NewDeserializer(factory ResourceFactory) *Deserializer {
	return &Deserializer{factory}
}

func (s *Deserializer) Deserialize(ty ResourceType, rawList []cty.Value) ([]Resource, error) {
	resources := make([]Resource, 0)
	for _, rawResource := range rawList {
		rawResource := rawResource
		var attrs Attributes

		bytes, _ := ctyjson.Marshal(rawResource, rawResource.Type())
		err := json.Unmarshal(bytes, &attrs)
		if err != nil {
			return nil, err
		}

		res := s.factory.CreateAbstractResource(ty.String(), rawResource.GetAttr("id").AsString(), attrs)
		resources = append(resources, res)
	}
	return resources, nil
}
