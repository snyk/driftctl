package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type NatGatewayDeserializer struct {
}

func NewNatGatewayDeserializer() *NatGatewayDeserializer {
	return &NatGatewayDeserializer{}
}

func (s *NatGatewayDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsNatGatewayResourceType
}

func (s NatGatewayDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range rawList {
		rawResource := rawResource
		resource, err := decodeNatGateway(&rawResource)
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

func decodeNatGateway(raw *cty.Value) (*resourceaws.AwsNatGateway, error) {
	var decoded resourceaws.AwsNatGateway
	if err := gocty.FromCtyValue(*raw, &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}
