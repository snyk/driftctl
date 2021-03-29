package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type IamRoleDeserializer struct {
}

func NewIamRoleDeserializer() *IamRoleDeserializer {
	return &IamRoleDeserializer{}
}

func (s IamRoleDeserializer) HandledType() resource.ResourceType {
	return aws.AwsIamRoleResourceType
}

func (s IamRoleDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, raw := range rawList {
		res, err := decodeIamRole(raw)
		if err != nil {
			logrus.Warnf("error when reading iam role %s : %+v", res, err)
			return nil, err
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func decodeIamRole(raw cty.Value) (resource.Resource, error) {
	var decoded aws.AwsIamRole
	if err := gocty.FromCtyValue(raw, &decoded); err != nil {
		return nil, err
	}
	decoded.CtyVal = &raw
	return &decoded, nil
}
