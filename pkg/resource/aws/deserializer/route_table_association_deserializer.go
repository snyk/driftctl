package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type RouteTableAssociationDeserializer struct {
}

func NewRouteTableAssociationDeserializer() *RouteTableAssociationDeserializer {
	return &RouteTableAssociationDeserializer{}
}

func (s *RouteTableAssociationDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsRouteTableAssociationResourceType
}

func (s RouteTableAssociationDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range rawList {
		resource, err := decodeRouteTableAssociation(&rawResource)
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

func decodeRouteTableAssociation(raw *cty.Value) (*resourceaws.AwsRouteTableAssociation, error) {
	var decoded resourceaws.AwsRouteTableAssociation
	if err := gocty.FromCtyValue(*raw, &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}
