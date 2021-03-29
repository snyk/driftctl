package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type IamPolicyDeserializer struct {
}

func NewIamPolicyDeserializer() *IamPolicyDeserializer {
	return &IamPolicyDeserializer{}
}

func (s IamPolicyDeserializer) HandledType() resource.ResourceType {
	return aws.AwsIamPolicyResourceType
}

func (s IamPolicyDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, raw := range rawList {
		res, err := decodeIamPolicy(raw)
		if err != nil {
			logrus.Warnf("error when reading iam policy %s : %+v", res, err)
			return nil, err
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func decodeIamPolicy(raw cty.Value) (resource.Resource, error) {
	var decoded aws.AwsIamPolicy
	if err := gocty.FromCtyValue(raw, &decoded); err != nil {
		return nil, err
	}
	decoded.CtyVal = &raw
	return &decoded, nil
}
