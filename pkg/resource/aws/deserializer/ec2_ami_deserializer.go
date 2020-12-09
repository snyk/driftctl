package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type EC2AmiDeserializer struct {
}

func NewEC2AmiDeserializer() *EC2AmiDeserializer {
	return &EC2AmiDeserializer{}
}

func (s EC2AmiDeserializer) HandledType() resource.ResourceType {
	return aws.AwsAmiResourceType
}

func (s EC2AmiDeserializer) Deserialize(imageList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawAmi := range imageList {
		image, err := decodeEC2Ami(rawAmi)
		if err != nil {
			logrus.Warnf("error when reading image %s : %+v", image, err)
			return nil, err
		}
		resources = append(resources, image)
	}
	return resources, nil
}

func decodeEC2Ami(rawAmi cty.Value) (resource.Resource, error) {
	var decodedImage aws.AwsAmi
	if err := gocty.FromCtyValue(rawAmi, &decodedImage); err != nil {
		return nil, err
	}
	return &decodedImage, nil
}
