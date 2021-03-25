package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type SubnetDeserializer struct {
}

func NewSubnetDeserializer() *SubnetDeserializer {
	return &SubnetDeserializer{}
}

func (s *SubnetDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsSubnetResourceType
}

func (s SubnetDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range rawList {
		rawResource := rawResource
		resource, err := decodeSubnet(&rawResource)
		if err != nil {
			logrus.Warnf("Error when deserializing resource %+v : %+v", rawResource, err)
			return nil, err
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func decodeSubnet(raw *cty.Value) (*resourceaws.AwsSubnet, error) {
	var decoded resourceaws.AwsSubnet
	if err := gocty.FromCtyValue(*raw, &decoded); err != nil {
		return nil, err
	}
	decoded.CtyVal = raw
	return &decoded, nil
}
