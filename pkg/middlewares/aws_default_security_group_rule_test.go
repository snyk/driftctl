package middlewares

import (
	"reflect"
	"testing"

	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

func TestAwsDefaultSecurityGroupRule_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    *[]*resource.Resource
		resourcesFromState *[]*resource.Resource
		expected           *[]*resource.Resource
		wantErr            bool
	}{
		{
			name: "Should ignore default rules if not managed",
			remoteResources: &[]*resource.Resource{
				{
					Id:    "default-sg",
					Type:  aws.AwsDefaultSecurityGroupResourceType,
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:   "default-ingress",
					Type: aws.AwsSecurityGroupRuleResourceType,
					Attrs: &resource.Attributes{
						"type":                     "ingress",
						"from_port":                float64(0),
						"to_port":                  float64(0),
						"protocol":                 "-1",
						"security_group_id":        "default-sg",
						"source_security_group_id": "default-sg",
						"self":                     true,
					},
				},
				&resource.Resource{
					Id:   "default-egress",
					Type: aws.AwsSecurityGroupRuleResourceType,
					Attrs: &resource.Attributes{
						"type":              "egress",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"protocol":          "-1",
						"security_group_id": "default-sg",
						"self":              false,
						"cidr_blocks":       []interface{}{"0.0.0.0/0"},
					},
				},
				&resource.Resource{
					Id:   "dummy-ingress",
					Type: aws.AwsSecurityGroupRuleResourceType,
					Attrs: &resource.Attributes{
						"type":              "ingress",
						"from_port":         float64(22),
						"to_port":           float64(22),
						"protocol":          "tcp",
						"security_group_id": "default-sg",
						"cidr_blocks":       []interface{}{"1.2.3.4/32"},
					},
				},
				&resource.Resource{
					Id:   "dummy-egress",
					Type: aws.AwsSecurityGroupRuleResourceType,
					Attrs: &resource.Attributes{
						"type":              "egress",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"protocol":          "-1",
						"security_group_id": "default-sg",
						"self":              false,
						"ipv6_cidr_blocks":  []interface{}{"::/0"},
					},
				},
			},
			resourcesFromState: &[]*resource.Resource{},
			expected: &[]*resource.Resource{
				&resource.Resource{
					Id:    "default-sg",
					Type:  aws.AwsDefaultSecurityGroupResourceType,
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:   "dummy-ingress",
					Type: aws.AwsSecurityGroupRuleResourceType,
					Attrs: &resource.Attributes{
						"type":              "ingress",
						"from_port":         float64(22),
						"to_port":           float64(22),
						"protocol":          "tcp",
						"security_group_id": "default-sg",
						"cidr_blocks":       []interface{}{"1.2.3.4/32"},
					},
				},
				&resource.Resource{
					Id:   "dummy-egress",
					Type: aws.AwsSecurityGroupRuleResourceType,
					Attrs: &resource.Attributes{
						"type":              "egress",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"protocol":          "-1",
						"security_group_id": "default-sg",
						"self":              false,
						"ipv6_cidr_blocks":  []interface{}{"::/0"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Should not ignore default rules if managed",
			remoteResources: &[]*resource.Resource{
				&resource.Resource{
					Id:    "default-sg",
					Type:  aws.AwsDefaultSecurityGroupResourceType,
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:    "dummy-sg",
					Type:  aws.AwsSecurityGroupResourceType,
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:   "default-ingress",
					Type: aws.AwsSecurityGroupRuleResourceType,
					Attrs: &resource.Attributes{
						"type":                     "ingress",
						"from_port":                float64(0),
						"to_port":                  float64(0),
						"protocol":                 "-1",
						"security_group_id":        "default-sg",
						"source_security_group_id": "default-sg",
						"self":                     true,
					},
				},
				&resource.Resource{
					Id:   "default-egress",
					Type: aws.AwsSecurityGroupRuleResourceType,
					Attrs: &resource.Attributes{
						"type":              "egress",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"protocol":          "-1",
						"security_group_id": "default-sg",
						"self":              false,
						"cidr_blocks":       []interface{}{"0.0.0.0/0"},
					},
				},
				&resource.Resource{
					Id:   "dummy-ingress",
					Type: aws.AwsSecurityGroupRuleResourceType,
					Attrs: &resource.Attributes{
						"type":              "ingress",
						"from_port":         float64(22),
						"to_port":           float64(22),
						"protocol":          "tcp",
						"security_group_id": "dummy-sg",
						"cidr_blocks":       []interface{}{"1.2.3.4/32"},
					},
				},
				&resource.Resource{
					Id:   "dummy-egress",
					Type: aws.AwsSecurityGroupRuleResourceType,
					Attrs: &resource.Attributes{
						"type":              "egress",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"protocol":          "-1",
						"security_group_id": "dummy-sg",
						"self":              false,
						"ipv6_cidr_blocks":  []interface{}{"::/0"},
					},
				},
				&resource.Resource{
					Id:   "default-egress-2",
					Type: aws.AwsSecurityGroupRuleResourceType,
					Attrs: &resource.Attributes{
						"type":              "egress",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"protocol":          "-1",
						"security_group_id": "dummy-sg",
						"self":              false,
						"cidr_blocks":       []interface{}{"0.0.0.0/0"},
					},
				},
			},
			resourcesFromState: &[]*resource.Resource{
				&resource.Resource{
					Id:    "default-sg",
					Type:  aws.AwsDefaultSecurityGroupResourceType,
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:   "default-ingress",
					Type: aws.AwsSecurityGroupRuleResourceType,
					Attrs: &resource.Attributes{
						"type":                     "ingress",
						"from_port":                float64(0),
						"to_port":                  float64(0),
						"protocol":                 "-1",
						"security_group_id":        "default-sg",
						"source_security_group_id": "default-sg",
						"self":                     true,
					},
				},
				&resource.Resource{
					Id:   "default-egress",
					Type: aws.AwsSecurityGroupRuleResourceType,
					Attrs: &resource.Attributes{
						"type":              "egress",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"protocol":          "-1",
						"security_group_id": "default-sg",
						"self":              false,
						"cidr_blocks":       []interface{}{"0.0.0.0/0"},
					},
				},
			},
			expected: &[]*resource.Resource{
				&resource.Resource{
					Id:    "default-sg",
					Type:  aws.AwsDefaultSecurityGroupResourceType,
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:    "dummy-sg",
					Type:  aws.AwsSecurityGroupResourceType,
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:   "default-ingress",
					Type: aws.AwsSecurityGroupRuleResourceType,
					Attrs: &resource.Attributes{
						"type":                     "ingress",
						"from_port":                float64(0),
						"to_port":                  float64(0),
						"protocol":                 "-1",
						"security_group_id":        "default-sg",
						"source_security_group_id": "default-sg",
						"self":                     true,
					},
				},
				&resource.Resource{
					Id:   "default-egress",
					Type: aws.AwsSecurityGroupRuleResourceType,
					Attrs: &resource.Attributes{
						"type":              "egress",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"protocol":          "-1",
						"security_group_id": "default-sg",
						"self":              false,
						"cidr_blocks":       []interface{}{"0.0.0.0/0"},
					},
				},
				&resource.Resource{
					Id:   "dummy-ingress",
					Type: aws.AwsSecurityGroupRuleResourceType,
					Attrs: &resource.Attributes{
						"type":              "ingress",
						"from_port":         float64(22),
						"to_port":           float64(22),
						"protocol":          "tcp",
						"security_group_id": "dummy-sg",
						"cidr_blocks":       []interface{}{"1.2.3.4/32"},
					},
				},
				&resource.Resource{
					Id:   "dummy-egress",
					Type: aws.AwsSecurityGroupRuleResourceType,
					Attrs: &resource.Attributes{
						"type":              "egress",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"protocol":          "-1",
						"security_group_id": "dummy-sg",
						"self":              false,
						"ipv6_cidr_blocks":  []interface{}{"::/0"},
					},
				},
				&resource.Resource{
					Id:   "default-egress-2",
					Type: aws.AwsSecurityGroupRuleResourceType,
					Attrs: &resource.Attributes{
						"type":              "egress",
						"from_port":         float64(0),
						"to_port":           float64(0),
						"protocol":          "-1",
						"security_group_id": "dummy-sg",
						"self":              false,
						"cidr_blocks":       []interface{}{"0.0.0.0/0"},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := AwsDefaultSecurityGroupRule{}
			if err := m.Execute(tt.remoteResources, tt.resourcesFromState); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.remoteResources, tt.expected) {
				t.Fatalf("Expected results mismatch")
			}
		})
	}
}
