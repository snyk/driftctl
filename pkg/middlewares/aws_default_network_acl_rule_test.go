package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

func TestAwsDefaultNetworkACLRule_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []*resource.Resource
		resourcesFromState []*resource.Resource
		expected           []*resource.Resource
	}{
		{
			"default network ACL rule is not ignored when managed by IaC",
			[]*resource.Resource{
				{
					Id: "fake",
				},
				{
					Id:   "default-acl-rule",
					Type: aws.AwsNetworkACLRuleResourceType,
					Attrs: &resource.Attributes{
						"rule_number": float64(32767),
						"rule_action": "deny",
						"cidr_block":  "0.0.0.0/0",
						"protocol":    "-1",
					},
				},
				{
					Id:   "non-default-acl",
					Type: aws.AwsNetworkACLRuleResourceType,
					Attrs: &resource.Attributes{
						"rule_number": float64(100),
					},
				},
				{
					Id:   "non-default-acl-2",
					Type: aws.AwsNetworkACLRuleResourceType,
					Attrs: &resource.Attributes{
						"rule_number": float64(32767),
						"rule_action": "allow",
					},
				},
				{
					Id:   "non-default-acl-3",
					Type: aws.AwsNetworkACLRuleResourceType,
					Attrs: &resource.Attributes{
						"rule_number": float64(32767),
						"rule_action": "deny",
						"cidr_block":  "1.2.3.0/0",
					},
				},
				{
					Id:   "non-default-acl-4",
					Type: aws.AwsNetworkACLRuleResourceType,
					Attrs: &resource.Attributes{
						"rule_number": float64(32767),
						"rule_action": "deny",
						"cidr_block":  "0.0.0.0/0",
						"protocol":    "6",
					},
				},
			},
			[]*resource.Resource{
				{
					Id:   "default-acl-rule",
					Type: aws.AwsNetworkACLRuleResourceType,
				},
			},
			[]*resource.Resource{
				{
					Id: "fake",
				},
				{
					Id:   "default-acl-rule",
					Type: aws.AwsNetworkACLRuleResourceType,
					Attrs: &resource.Attributes{
						"rule_number": float64(32767),
						"rule_action": "deny",
						"cidr_block":  "0.0.0.0/0",
						"protocol":    "-1",
					},
				},
				{
					Id:   "non-default-acl",
					Type: aws.AwsNetworkACLRuleResourceType,
					Attrs: &resource.Attributes{
						"rule_number": float64(100),
					},
				},
				{
					Id:   "non-default-acl-2",
					Type: aws.AwsNetworkACLRuleResourceType,
					Attrs: &resource.Attributes{
						"rule_number": float64(32767),
						"rule_action": "allow",
					},
				},
				{
					Id:   "non-default-acl-3",
					Type: aws.AwsNetworkACLRuleResourceType,
					Attrs: &resource.Attributes{
						"rule_number": float64(32767),
						"rule_action": "deny",
						"cidr_block":  "1.2.3.0/0",
					},
				},
				{
					Id:   "non-default-acl-4",
					Type: aws.AwsNetworkACLRuleResourceType,
					Attrs: &resource.Attributes{
						"rule_number": float64(32767),
						"rule_action": "deny",
						"cidr_block":  "0.0.0.0/0",
						"protocol":    "6",
					},
				},
			},
		},
		{
			"default network acl rule is ignored when not managed by IaC",
			[]*resource.Resource{
				{
					Id: "fake",
				},
				{
					Id:   "default-acl-rule",
					Type: aws.AwsNetworkACLRuleResourceType,
					Attrs: &resource.Attributes{
						"network_acl_id": "my-network",
						"rule_number":    float64(32767),
						"rule_action":    "deny",
						"cidr_block":     "0.0.0.0/0",
						"protocol":       "-1",
					},
				},
				{
					Id:   "default-ipv6-acl-rule",
					Type: aws.AwsNetworkACLRuleResourceType,
					Attrs: &resource.Attributes{
						"network_acl_id":  "my-network",
						"rule_number":     float64(32768),
						"rule_action":     "deny",
						"ipv6_cidr_block": "::/0",
						"protocol":        "-1",
					},
				},
				{
					Id:   "non-default-acl",
					Type: aws.AwsNetworkACLRuleResourceType,
					Attrs: &resource.Attributes{
						"rule_number": float64(32767),
						"rule_action": "deny",
						"cidr_block":  "0.0.0.0/0",
						"protocol":    "6",
					},
				},
			},
			[]*resource.Resource{},
			[]*resource.Resource{
				{
					Id: "fake",
				},
				{
					Id:   "non-default-acl",
					Type: aws.AwsNetworkACLRuleResourceType,
					Attrs: &resource.Attributes{
						"rule_number": float64(32767),
						"rule_action": "deny",
						"cidr_block":  "0.0.0.0/0",
						"protocol":    "6",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsDefaultNetworkACLRule()
			err := m.Execute(&tt.remoteResources, &tt.resourcesFromState)
			if err != nil {
				t.Fatal(err)
			}
			changelog, err := diff.Diff(tt.expected, tt.remoteResources)
			if err != nil {
				t.Fatal(err)
			}
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), awsutil.Prettify(change.From), awsutil.Prettify(change.To))
				}
			}
		})
	}
}
