package deserializer

import (
	"fmt"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// WARN this deserializer will also mutate the return type to PolicyAttachment
type IamRolePolicyAttachmentDeserializer struct {
}

func NewIamRolePolicyAttachmentDeserializer() *IamRolePolicyAttachmentDeserializer {
	return &IamRolePolicyAttachmentDeserializer{}
}

func (s IamRolePolicyAttachmentDeserializer) HandledType() resource.ResourceType {
	return aws.AwsIamRolePolicyAttachmentResourceType
}

func (s IamRolePolicyAttachmentDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, raw := range rawList {
		rolePolicyAttachment, err := decodeIamRolePolicyAttachment(raw)
		if err != nil {
			logrus.Warnf("error when deserializing iam role policy attachment %s : %+v", rawList, err)
			return nil, err
		}
		policyAttachment := aws.AwsIamPolicyAttachment{
			Id:        fmt.Sprintf("%s-%s", *rolePolicyAttachment.Role, *rolePolicyAttachment.PolicyArn), // generate unique id
			Name:      awssdk.String(rolePolicyAttachment.Id),
			PolicyArn: rolePolicyAttachment.PolicyArn,
			Users:     &[]string{},
			Groups:    &[]string{},
			Roles:     &[]string{*rolePolicyAttachment.Role},
		}
		resources = append(resources, &policyAttachment)
	}
	return resources, nil
}

func decodeIamRolePolicyAttachment(raw cty.Value) (*aws.AwsIamRolePolicyAttachment, error) {
	var decoded aws.AwsIamRolePolicyAttachment
	if err := gocty.FromCtyValue(raw, &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}
