package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type RouteTableDeserializer struct {
}

func NewRouteTableDeserializer() *RouteTableDeserializer {
	return &RouteTableDeserializer{}
}

func (s *RouteTableDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsRouteTableResourceType
}

func (s RouteTableDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range rawList {
		rawResource := rawResource
		resource, err := decodeRouteTable(&rawResource)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type": s.HandledType(),
			}).Warnf("Error when deserializing resource %+v : %+v", rawResource, err)
			return nil, err
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func decodeRouteTable(raw *cty.Value) (*resourceaws.AwsRouteTable, error) {
	var decoded resourceaws.AwsRouteTable
	if err := gocty.FromCtyValue(*raw, &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}
