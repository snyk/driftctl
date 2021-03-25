package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type VPCSecurityGroupRuleDeserializer struct {
}

func NewVPCSecurityGroupRuleDeserializer() *VPCSecurityGroupRuleDeserializer {
	return &VPCSecurityGroupRuleDeserializer{}
}

func (s VPCSecurityGroupRuleDeserializer) HandledType() resource.ResourceType {
	return aws.AwsSecurityGroupRuleResourceType
}

func (s VPCSecurityGroupRuleDeserializer) Deserialize(sgRuleList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawSecurityGroupRule := range sgRuleList {
		sgRule, err := decodeVPCSecurityGroupRule(rawSecurityGroupRule)
		if err != nil {
			logrus.Warnf("error when reading security group rule %s : %+v", sgRule, err)
			return nil, err
		}
		resources = append(resources, sgRule)
	}
	return resources, nil
}

func decodeVPCSecurityGroupRule(rawSecurityGroupRule cty.Value) (resource.Resource, error) {
	var decodedSgRule aws.AwsSecurityGroupRule
	if err := gocty.FromCtyValue(rawSecurityGroupRule, &decodedSgRule); err != nil {
		return nil, err
	}
	decodedSgRule.CtyVal = &rawSecurityGroupRule
	return &decodedSgRule, nil
}
