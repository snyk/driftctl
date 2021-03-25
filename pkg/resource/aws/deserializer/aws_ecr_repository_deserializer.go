package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type ECRRepositoryDeserializer struct {
}

func NewECRRepositoryDeserializer() *ECRRepositoryDeserializer {
	return &ECRRepositoryDeserializer{}
}

func (s *ECRRepositoryDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsEcrRepositoryResourceType
}

func (s ECRRepositoryDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range rawList {
		rawResource := rawResource
		resource, err := decodeECRRepository(&rawResource)
		if err != nil {
			logrus.Warnf("Error when deserializing resource %+v : %+v", rawResource, err)
			return nil, err
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func decodeECRRepository(raw *cty.Value) (*resourceaws.AwsEcrRepository, error) {
	var decoded resourceaws.AwsEcrRepository
	if err := gocty.FromCtyValue(*raw, &decoded); err != nil {
		return nil, err
	}
	decoded.CtyVal = raw
	return &decoded, nil
}
