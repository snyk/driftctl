package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type IamAccessKeyDeserializer struct {
}

func NewIamAccessKeyDeserializer() *IamAccessKeyDeserializer {
	return &IamAccessKeyDeserializer{}
}

func (s IamAccessKeyDeserializer) HandledType() resource.ResourceType {
	return aws.AwsIamAccessKeyResourceType
}

func (s IamAccessKeyDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, raw := range rawList {
		res, err := decodeIamAccessKey(raw)
		if err != nil {
			logrus.Warnf("error when reading iam access key %s : %+v", res, err)
			return nil, err
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func decodeIamAccessKey(raw cty.Value) (resource.Resource, error) {
	var decoded aws.AwsIamAccessKey
	if err := gocty.FromCtyValue(raw, &decoded); err != nil {
		return nil, err
	}
	decoded.CtyVal = &raw
	return &decoded, nil
}
