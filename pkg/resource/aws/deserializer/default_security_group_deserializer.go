package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/zclconf/go-cty/cty"
)

type DefaultSecurityGroupDeserializer struct {
	deserializer
}

func NewDefaultSecurityGroupDeserializerForState() *DefaultSecurityGroupDeserializer {
	return &DefaultSecurityGroupDeserializer{
		deserializer{
			normalize: func(res resource.Resource) error {
				r := res.(*resourceaws.AwsDefaultSecurityGroup)
				if r.Ingress != nil {
					r.Ingress = nil
				}
				if r.Egress != nil {
					r.Egress = nil
				}
				return nil
			},
		},
	}
}

func NewDefaultSecurityGroupDeserializerForProvider() *DefaultSecurityGroupDeserializer {
	return &DefaultSecurityGroupDeserializer{
		deserializer{
			normalize: func(res resource.Resource) error {
				r := res.(*resourceaws.AwsDefaultSecurityGroup)
				if r.Ingress != nil {
					r.Ingress = nil
				}
				if r.Egress != nil {
					r.Egress = nil
				}
				return nil
			},
		},
	}
}

func (s *DefaultSecurityGroupDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsDefaultSecurityGroupResourceType
}

func (s DefaultSecurityGroupDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	return s.deserialize(rawList, &resourceaws.AwsDefaultSecurityGroup{})
}
