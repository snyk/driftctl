package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type EC2EipDeserializer struct {
}

func NewEC2EipDeserializer() *EC2EipDeserializer {
	return &EC2EipDeserializer{}
}

func (s EC2EipDeserializer) HandledType() resource.ResourceType {
	return aws.AwsEipResourceType
}

func (s EC2EipDeserializer) Deserialize(addressList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawAddress := range addressList {
		address, err := decodeEC2Eip(rawAddress)
		if err != nil {
			logrus.Warnf("error when reading eip %s : %+v", address, err)
			return nil, err
		}
		resources = append(resources, address)
	}
	return resources, nil
}

func decodeEC2Eip(rawAddress cty.Value) (resource.Resource, error) {
	var decodedAddress aws.AwsEip
	if err := gocty.FromCtyValue(rawAddress, &decodedAddress); err != nil {
		return nil, err
	}
	return &decodedAddress, nil
}
