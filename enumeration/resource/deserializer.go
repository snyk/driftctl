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

func (s *Deserializer) Deserialize(ty string, rawList []cty.Value) ([]*Resource, error) {
	resources := make([]*Resource, 0)
	for _, rawResource := range rawList {
		rawResource := rawResource
		res, err := s.DeserializeOne(ty, rawResource)
		if err != nil {
			return nil, err
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func (s *Deserializer) DeserializeOne(ty string, value cty.Value) (*Resource, error) {
	if value.IsNull() {
		return nil, nil
	}

	// Marked values cannot be deserialized to JSON.
	// For example, this ensures we can deserialize sensitive values too.
	unmarkedVal, _ := value.UnmarkDeep()

	var attrs Attributes
	bytes, _ := ctyjson.Marshal(unmarkedVal, unmarkedVal.Type())
	err := json.Unmarshal(bytes, &attrs)
	if err != nil {
		return nil, err
	}

	return s.factory.CreateAbstractResource(ty, value.GetAttr("id").AsString(), attrs), nil
}
