package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type VPCSecurityGroupDeserializer struct {
}

func NewVPCSecurityGroupDeserializer() *VPCSecurityGroupDeserializer {
	return &VPCSecurityGroupDeserializer{}
}

func (s VPCSecurityGroupDeserializer) HandledType() resource.ResourceType {
	return aws.AwsSecurityGroupResourceType
}

func (s VPCSecurityGroupDeserializer) Deserialize(sgList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawSecurityGroup := range sgList {
		sg, err := decodeVPCSecurityGroup(rawSecurityGroup)
		if err != nil {
			logrus.Warnf("error when reading security group %s : %+v", sg, err)
			return nil, err
		}
		resources = append(resources, sg)
	}
	return resources, nil
}

func decodeVPCSecurityGroup(rawSecurityGroup cty.Value) (resource.Resource, error) {
	var decodedSg aws.AwsSecurityGroup
	if err := gocty.FromCtyValue(rawSecurityGroup, &decodedSg); err != nil {
		return nil, err
	}
	return &decodedSg, nil
}
