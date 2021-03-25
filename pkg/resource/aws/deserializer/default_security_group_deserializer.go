package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type DefaultSecurityGroupDeserializer struct {
}

func NewDefaultSecurityGroupDeserializer() *DefaultSecurityGroupDeserializer {
	return &DefaultSecurityGroupDeserializer{}
}

func (s *DefaultSecurityGroupDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsDefaultSecurityGroupResourceType
}

func (s DefaultSecurityGroupDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range rawList {
		rawResource := rawResource
		resource, err := decodeDefaultSecurityGroup(&rawResource)
		if err != nil {
			logrus.Warnf("Error when deserializing resource %+v : %+v", rawResource, err)
			return nil, err
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func decodeDefaultSecurityGroup(raw *cty.Value) (*resourceaws.AwsDefaultSecurityGroup, error) {
	var decoded resourceaws.AwsDefaultSecurityGroup
	if err := gocty.FromCtyValue(*raw, &decoded); err != nil {
		return nil, err
	}
	decoded.CtyVal = raw
	return &decoded, nil
}
