package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type DefaultVPCDeserializer struct {
}

func NewDefaultVPCDeserializer() *DefaultVPCDeserializer {
	return &DefaultVPCDeserializer{}
}

func (s *DefaultVPCDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsDefaultVpcResourceType
}

func (s DefaultVPCDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range rawList {
		rawResource := rawResource
		resource, err := decodeDefaultVPC(&rawResource)
		if err != nil {
			logrus.Warnf("Error when deserializing resource %+v : %+v", rawResource, err)
			return nil, err
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func decodeDefaultVPC(raw *cty.Value) (*resourceaws.AwsDefaultVpc, error) {
	var decoded resourceaws.AwsDefaultVpc
	if err := gocty.FromCtyValue(*raw, &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}
