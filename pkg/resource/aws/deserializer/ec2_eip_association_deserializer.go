package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type EC2EipAssociationDeserializer struct {
}

func NewEC2EipAssociationDeserializer() *EC2EipAssociationDeserializer {
	return &EC2EipAssociationDeserializer{}
}

func (s EC2EipAssociationDeserializer) HandledType() resource.ResourceType {
	return aws.AwsEipAssociationResourceType
}

func (s EC2EipAssociationDeserializer) Deserialize(assocList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawAssoc := range assocList {
		assoc, err := decodeEC2EipAssociation(rawAssoc)
		if err != nil {
			logrus.Warnf("error when reading eip association %s : %+v", assoc, err)
			return nil, err
		}
		resources = append(resources, assoc)
	}
	return resources, nil
}

func decodeEC2EipAssociation(rawAssoc cty.Value) (resource.Resource, error) {
	var decodedAssoc aws.AwsEipAssociation
	if err := gocty.FromCtyValue(rawAssoc, &decodedAssoc); err != nil {
		return nil, err
	}
	decodedAssoc.CtyVal = &rawAssoc
	return &decodedAssoc, nil
}
