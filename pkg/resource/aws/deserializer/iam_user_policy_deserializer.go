package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type IamUserPolicyDeserializer struct {
}

func NewIamUserPolicyDeserializer() *IamUserPolicyDeserializer {
	return &IamUserPolicyDeserializer{}
}

func (s IamUserPolicyDeserializer) HandledType() resource.ResourceType {
	return aws.AwsIamUserPolicyResourceType
}

func (s IamUserPolicyDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, raw := range rawList {
		res, err := decodeIamUserPolicy(raw)
		if err != nil {
			logrus.Warnf("error when deserializing iam user policy %s : %+v", raw, err)
			return nil, err
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func decodeIamUserPolicy(raw cty.Value) (resource.Resource, error) {
	var decoded aws.AwsIamUserPolicy
	if err := gocty.FromCtyValue(raw, &decoded); err != nil {
		return nil, err
	}
	decoded.CtyVal = &raw
	return &decoded, nil
}
