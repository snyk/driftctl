package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type KMSAliasDeserializer struct {
}

func NewKMSAliasDeserializer() *KMSAliasDeserializer {
	return &KMSAliasDeserializer{}
}

func (s *KMSAliasDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsKmsAliasResourceType
}

func (s KMSAliasDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range rawList {
		rawResource := rawResource
		resource, err := decodeKMSAlias(&rawResource)
		if err != nil {
			logrus.Warnf("Error when deserializing resource %+v : %+v", rawResource, err)
			return nil, err
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func decodeKMSAlias(raw *cty.Value) (*resourceaws.AwsKmsAlias, error) {
	var decoded resourceaws.AwsKmsAlias
	if err := gocty.FromCtyValue(*raw, &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}
