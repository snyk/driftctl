package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type EC2KeyPairDeserializer struct {
}

func NewEC2KeyPairDeserializer() *EC2KeyPairDeserializer {
	return &EC2KeyPairDeserializer{}
}

func (s EC2KeyPairDeserializer) HandledType() resource.ResourceType {
	return aws.AwsKeyPairResourceType
}

func (s EC2KeyPairDeserializer) Deserialize(kpList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawKeyPair := range kpList {
		kp, err := decodeEC2KeyPair(rawKeyPair)
		if err != nil {
			logrus.Warnf("error when reading key pair %s : %+v", kp, err)
			return nil, err
		}
		resources = append(resources, kp)
	}
	return resources, nil
}

func decodeEC2KeyPair(rawKeyPair cty.Value) (resource.Resource, error) {
	var decodedKp aws.AwsKeyPair
	if err := gocty.FromCtyValue(rawKeyPair, &decodedKp); err != nil {
		return nil, err
	}
	return &decodedKp, nil
}
