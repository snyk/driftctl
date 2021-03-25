package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type DefaultRouteTableDeserializer struct {
}

func NewDefaultRouteTableDeserializer() *DefaultRouteTableDeserializer {
	return &DefaultRouteTableDeserializer{}
}

func (s *DefaultRouteTableDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsDefaultRouteTableResourceType
}

func (s DefaultRouteTableDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range rawList {
		rawResource := rawResource
		resource, err := decodeDefaultRouteTable(&rawResource)
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

func decodeDefaultRouteTable(raw *cty.Value) (*resourceaws.AwsDefaultRouteTable, error) {
	var decoded resourceaws.AwsDefaultRouteTable
	if err := gocty.FromCtyValue(*raw, &decoded); err != nil {
		return nil, err
	}
	decoded.CtyVal = raw
	return &decoded, nil
}
