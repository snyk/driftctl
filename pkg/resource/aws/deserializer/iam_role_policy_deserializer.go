package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type IamRolePolicyDeserializer struct {
}

func NewIamRolePolicyDeserializer() *IamRolePolicyDeserializer {
	return &IamRolePolicyDeserializer{}
}

func (s IamRolePolicyDeserializer) HandledType() resource.ResourceType {
	return aws.AwsIamRolePolicyResourceType
}

func (s IamRolePolicyDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, raw := range rawList {
		res, err := decodeIamRolePolicy(raw)
		if err != nil {
			logrus.Warnf("error when reading iam role policy %s : %+v", res, err)
			return nil, err
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func decodeIamRolePolicy(raw cty.Value) (resource.Resource, error) {
	var decoded aws.AwsIamRolePolicy
	if err := gocty.FromCtyValue(raw, &decoded); err != nil {
		return nil, err
	}
	decoded.CtyVal = &raw
	return &decoded, nil
}
