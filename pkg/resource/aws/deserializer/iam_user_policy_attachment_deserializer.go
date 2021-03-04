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
type IamUserPolicyAttachmentDeserializer struct {
}

func NewIamUserPolicyAttachmentDeserializer() *IamUserPolicyAttachmentDeserializer {
	return &IamUserPolicyAttachmentDeserializer{}
}

func (s IamUserPolicyAttachmentDeserializer) HandledType() resource.ResourceType {
	return aws.AwsIamUserPolicyAttachmentResourceType
}

func (s IamUserPolicyAttachmentDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, raw := range rawList {
		userPolicyAttachment, err := decodeIamUserPolicyAttachment(raw)
		if err != nil {
			logrus.Warnf("error when deserializing iam user policy attachment %s : %+v", rawList, err)
			return nil, err
		}
		policyAttachment := aws.AwsIamPolicyAttachment{
			Id:        fmt.Sprintf("%s-%s", *userPolicyAttachment.User, *userPolicyAttachment.PolicyArn), // generate unique id,
			Name:      awssdk.String(userPolicyAttachment.Id),
			PolicyArn: userPolicyAttachment.PolicyArn,
			Users:     &[]string{*userPolicyAttachment.User},
			Groups:    &[]string{},
			Roles:     &[]string{},
		}
		resources = append(resources, &policyAttachment)
	}
	return resources, nil
}

func decodeIamUserPolicyAttachment(raw cty.Value) (*aws.AwsIamUserPolicyAttachment, error) {
	var decoded aws.AwsIamUserPolicyAttachment
	if err := gocty.FromCtyValue(raw, &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}
