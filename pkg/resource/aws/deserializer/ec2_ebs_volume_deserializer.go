package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type EC2EbsVolumeDeserializer struct {
}

func NewEC2EbsVolumeDeserializer() *EC2EbsVolumeDeserializer {
	return &EC2EbsVolumeDeserializer{}
}

func (s EC2EbsVolumeDeserializer) HandledType() resource.ResourceType {
	return aws.AwsEbsVolumeResourceType
}

func (s EC2EbsVolumeDeserializer) Deserialize(volumeList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawVolume := range volumeList {
		volume, err := decodeEC2EbsVolume(rawVolume)
		if err != nil {
			logrus.Warnf("error when reading volume %s : %+v", volume, err)
			return nil, err
		}
		resources = append(resources, volume)
	}
	return resources, nil
}

func decodeEC2EbsVolume(rawVolume cty.Value) (resource.Resource, error) {
	var decodedVolume aws.AwsEbsVolume
	if err := gocty.FromCtyValue(rawVolume, &decodedVolume); err != nil {
		return nil, err
	}
	decodedVolume.CtyVal = &rawVolume
	return &decodedVolume, nil
}
