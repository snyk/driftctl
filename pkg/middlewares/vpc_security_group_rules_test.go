package middlewares

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
	"github.com/snyk/driftctl/pkg/terraform"
)

func TestVPCSecurityGroupRuleSanitizer(t *testing.T) {

	factory := &terraform.MockResourceFactory{}
	factory.On("CreateAbstractResource", aws.AwsSecurityGroupRuleResourceType, "sgrule-1175318309", mock.Anything).Times(1).Return(
		&resource.Resource{
			Id:    "sgrule-1175318309",
			Type:  aws.AwsSecurityGroupRuleResourceType,
			Attrs: &resource.Attributes{},
		}, nil)

	factory.On("CreateAbstractResource", aws.AwsSecurityGroupRuleResourceType, "sgrule-2582518759", mock.Anything).Times(1).Return(
		&resource.Resource{
			Id:    "sgrule-2582518759",
			Type:  aws.AwsSecurityGroupRuleResourceType,
			Attrs: &resource.Attributes{},
		}, nil)

	factory.On("CreateAbstractResource", aws.AwsSecurityGroupRuleResourceType, "sgrule-2165103420", mock.Anything).Times(1).Return(
		&resource.Resource{
			Id:    "sgrule-2165103420",
			Type:  aws.AwsSecurityGroupRuleResourceType,
			Attrs: &resource.Attributes{},
		}, nil)

	factory.On("CreateAbstractResource", aws.AwsSecurityGroupRuleResourceType, "sgrule-350400929", mock.Anything).Times(1).Return(
		&resource.Resource{
			Id:    "sgrule-350400929",
			Type:  aws.AwsSecurityGroupRuleResourceType,
			Attrs: &resource.Attributes{},
		}, nil)

	middleware := NewVPCSecurityGroupRuleSanitizer(factory)
	var remoteResources []*resource.Resource
	stateResources := []*resource.Resource{
		{
			Id:   "sg-test",
			Type: aws.AwsSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"id":   "sg-test",
				"name": "test",
			},
		},
		{
			Id:   "sgrule-3970541193",
			Type: aws.AwsSecurityGroupRuleResourceType,
			Attrs: &resource.Attributes{
				"id":                       "sgrule-3970541193",
				"type":                     "ingress",
				"security_group_id":        "sg-0254c038e32f25530",
				"protocol":                 "tcp",
				"from_port":                float64(0),
				"to_port":                  float64(65535),
				"self":                     true,
				"source_security_group_id": "sg-0254c038e32f25530",
			},
		},
		{
			Id:   "sgrule-845917806",
			Type: aws.AwsSecurityGroupRuleResourceType,
			Attrs: &resource.Attributes{
				"id":                "sgrule-845917806",
				"type":              "egress",
				"security_group_id": "sg-0cc8b3c3c2851705a",
				"protocol":          "-1",
				"from_port":         float64(0),
				"to_port":           float64(0),
				"cidr_blocks":       []interface{}{"0.0.0.0/0"},
				"ipv6_cidr_blocks":  []interface{}{"::/0"},
			},
		},
		{
			Id:   "sgrule-294318973",
			Type: aws.AwsSecurityGroupRuleResourceType,
			Attrs: &resource.Attributes{
				"id":                "sgrule-294318973",
				"type":              "ingress",
				"security_group_id": "sg-0254c038e32f25530",
				"protocol":          "-1",
				"from_port":         float64(0),
				"to_port":           float64(0),
				"cidr_blocks":       []interface{}{"1.2.0.0/16", "5.6.7.0/24"},
			},
		},
		{
			Id:   "sgrule-2471889226",
			Type: aws.AwsSecurityGroupRuleResourceType,
			Attrs: &resource.Attributes{
				"id":                "sgrule-2471889226",
				"type":              "ingress",
				"security_group_id": "sg-0254c038e32f25530",
				"protocol":          "tcp",
				"from_port":         float64(0),
				"to_port":           float64(0),
				"prefix_list_id":    []interface{}{"pl-abb451c2"},
			},
		},
		{
			Id:   "sgrule-3587309474",
			Type: aws.AwsSecurityGroupRuleResourceType,
			Attrs: &resource.Attributes{
				"id":                "sgrule-3587309474",
				"type":              "ingress",
				"security_group_id": "sg-0254c038e32f25530",
				"protocol":          "tcp",
				"from_port":         float64(0),
				"to_port":           float64(65535),
				"prefix_list_id":    []interface{}{"sg-9e0204ff"},
			},
		},
	}
	err := middleware.Execute(&remoteResources, &stateResources)
	if err != nil {
		t.Error(err)
	}
	if len(stateResources) != 8 {
		t.Error("Some security group rules were not split")
	}
}
