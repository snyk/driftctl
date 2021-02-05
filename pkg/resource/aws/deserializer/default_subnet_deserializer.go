package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type DefaultSubnetDeserializer struct {
}

func NewDefaultSubnetDeserializer() *DefaultSubnetDeserializer {
	return &DefaultSubnetDeserializer{}
}

func (s *DefaultSubnetDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsDefaultSubnetResourceType
}

func (s DefaultSubnetDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range rawList {
		rawResource := rawResource
		resource, err := decodeDefaultSubnet(&rawResource)
		if err != nil {
			logrus.Warnf("Error when deserializing resource %+v : %+v", rawResource, err)
			return nil, err
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func decodeDefaultSubnet(raw *cty.Value) (*resourceaws.AwsDefaultSubnet, error) {
	var decoded resourceaws.AwsDefaultSubnet
	if err := gocty.FromCtyValue(*raw, &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}
