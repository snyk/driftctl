package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type IamUserDeserializer struct {
}

func NewIamUserDeserializer() *IamUserDeserializer {
	return &IamUserDeserializer{}
}

func (s IamUserDeserializer) HandledType() resource.ResourceType {
	return aws.AwsIamUserResourceType
}

func (s IamUserDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, raw := range rawList {
		res, err := decodeIamUser(raw)
		if err != nil {
			logrus.Warnf("error when reading iam user %s : %+v", res, err)
			return nil, err
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func decodeIamUser(raw cty.Value) (resource.Resource, error) {
	var decoded aws.AwsIamUser
	if err := gocty.FromCtyValue(raw, &decoded); err != nil {
		return nil, err
	}
	decoded.CtyVal = &raw
	return &decoded, nil
}
