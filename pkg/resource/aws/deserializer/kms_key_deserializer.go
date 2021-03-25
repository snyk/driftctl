package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type KMSKeyDeserializer struct {
}

func NewKMSKeyDeserializer() *KMSKeyDeserializer {
	return &KMSKeyDeserializer{}
}

func (s *KMSKeyDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsKmsKeyResourceType
}

func (s KMSKeyDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range rawList {
		rawResource := rawResource
		resource, err := decodeKMSKey(&rawResource)
		if err != nil {
			logrus.Warnf("Error when deserializing resource %+v : %+v", rawResource, err)
			return nil, err
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func decodeKMSKey(raw *cty.Value) (*resourceaws.AwsKmsKey, error) {
	var decoded resourceaws.AwsKmsKey
	if err := gocty.FromCtyValue(*raw, &decoded); err != nil {
		return nil, err
	}
	decoded.CtyVal = raw
	return &decoded, nil
}
