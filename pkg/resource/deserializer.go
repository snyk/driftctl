package resource

import (
	"github.com/zclconf/go-cty/cty"
)

type Deserializer struct {
	factory ResourceFactory
}

func NewDeserializer(factory ResourceFactory) *Deserializer {
	return &Deserializer{factory}
}

func (s *Deserializer) Deserialize(ty string, rawList []cty.Value) ([]Resource, error) {
	resources := make([]Resource, 0)
	for _, rawResource := range rawList {
		rawResource := rawResource
		attrs := ToResourceAttributes(&rawResource)
		res := s.factory.CreateAbstractResource(ty, rawResource.GetAttr("id").AsString(), *attrs)
		resources = append(resources, res)
	}
	return resources, nil
}
