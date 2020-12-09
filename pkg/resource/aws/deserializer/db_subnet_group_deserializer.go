package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type DBSubnetGroupDeserializer struct {
}

func NewDBSubnetGroupDeserializer() *DBSubnetGroupDeserializer {
	return &DBSubnetGroupDeserializer{}
}

func (s *DBSubnetGroupDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsDbSubnetGroupResourceType
}

func (s DBSubnetGroupDeserializer) Deserialize(recordList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range recordList {
		resource, err := decodeDBSubnetGroup(&rawResource)
		if err != nil {
			logrus.Warnf("Error when deserializing resource %+v : %+v", rawResource, err)
			return nil, err
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func decodeDBSubnetGroup(raw *cty.Value) (*resourceaws.AwsDbSubnetGroup, error) {
	var decoded resourceaws.AwsDbSubnetGroup
	if err := gocty.FromCtyValue(*raw, &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}
