package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type IamPolicyAttachmentDeserializer struct {
}

func NewIamPolicyAttachmentDeserializer() *IamPolicyAttachmentDeserializer {
	return &IamPolicyAttachmentDeserializer{}
}

func (s IamPolicyAttachmentDeserializer) HandledType() resource.ResourceType {
	return aws.AwsIamPolicyAttachmentResourceType
}

func (s IamPolicyAttachmentDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, raw := range rawList {
		res, err := decodeIamPolicyAttachment(raw)
		if err != nil {
			logrus.Warnf("error when reading iam policy attachment %s : %+v", raw, err)
			return nil, err
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func decodeIamPolicyAttachment(raw cty.Value) (resource.Resource, error) {
	var decoded aws.AwsIamPolicyAttachment
	if err := gocty.FromCtyValue(raw, &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}
