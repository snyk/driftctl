package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type VPCDeserializer struct {
}

func NewVPCDeserializer() *VPCDeserializer {
	return &VPCDeserializer{}
}

func (s *VPCDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsVpcResourceType
}

func (s VPCDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range rawList {
		resource, err := decodeVPC(&rawResource)
		if err != nil {
			logrus.Warnf("Error when deserializing resource %+v : %+v", rawResource, err)
			return nil, err
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func decodeVPC(raw *cty.Value) (*resourceaws.AwsVpc, error) {
	var decoded resourceaws.AwsVpc
	if err := gocty.FromCtyValue(*raw, &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}
