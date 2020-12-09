package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type EC2InstanceDeserializer struct {
}

func NewEC2InstanceDeserializer() *EC2InstanceDeserializer {
	return &EC2InstanceDeserializer{}
}

func (s EC2InstanceDeserializer) HandledType() resource.ResourceType {
	return aws.AwsInstanceResourceType
}

func (s EC2InstanceDeserializer) Deserialize(instanceList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawInstance := range instanceList {
		instance, err := decodeEC2Instance(rawInstance)
		if err != nil {
			logrus.Warnf("error when reading instance %s : %+v", instance, err)
			return nil, err
		}
		resources = append(resources, instance)
	}
	return resources, nil
}

func decodeEC2Instance(rawInstance cty.Value) (resource.Resource, error) {
	var decodedInstance aws.AwsInstance
	if err := gocty.FromCtyValue(rawInstance, &decodedInstance); err != nil {
		return nil, err
	}
	return &decodedInstance, nil
}
