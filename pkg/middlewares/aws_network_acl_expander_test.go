package middlewares

import (
	"strings"
	"testing"

	"github.com/snyk/driftctl/enumeration/terraform"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

func TestAwsNetworkACLExpander_Execute(t *testing.T) {
	tests := []struct {
		name                                  string
		mock                                  func(factory *terraform.MockResourceFactory)
		remoteResources                       []*resource.Resource
		resourcesFromState                    []*resource.Resource
		expectedFromState, expectedFromRemote []*resource.Resource
	}{
		{
			name: "test nothing is expanded",
			remoteResources: []*resource.Resource{
				{
					Id: "fake",
				},
			},
			resourcesFromState: []*resource.Resource{
				{
					Id: "fake",
				},
				{
					Id:    "non-ingress-and-egress",
					Type:  aws.AwsNetworkACLResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			expectedFromState: []*resource.Resource{
				{
					Id: "fake",
				},
				{
					Id:    "non-ingress-and-egress",
					Type:  aws.AwsNetworkACLResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			expectedFromRemote: []*resource.Resource{
				{
					Id: "fake",
				},
			},
		},
		{
			name: "network ACL rule are expanded",
			remoteResources: []*resource.Resource{
				{
					Id: "fake",
				},
				{
					Id:   "ingress-and-egress-should-be-removed-from-remote-res",
					Type: aws.AwsNetworkACLResourceType,
					Attrs: &resource.Attributes{
						"ingress": "something",
						"egress":  "something",
					},
				},
			},
			mock: func(factory *terraform.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource",
					aws.AwsNetworkACLRuleResourceType,
					aws.CreateNetworkACLRuleID(
						"acl",
						100,
						false,
						"17",
					),
					map[string]interface{}{
						"network_acl_id":  "acl",
						"egress":          false,
						"rule_action":     "allow",
						"cidr_block":      "172.31.0.0/16",
						"from_port":       80,
						"icmp_code":       0,
						"icmp_type":       0,
						"ipv6_cidr_block": "",
						"protocol":        "17",
						"rule_number":     100,
						"to_port":         80,
					},
				).Once().Return(&resource.Resource{
					Id:   "acl-rule1",
					Type: aws.AwsNetworkACLRuleResourceType,
				})

				factory.On(
					"CreateAbstractResource",
					aws.AwsNetworkACLRuleResourceType,
					aws.CreateNetworkACLRuleID(
						"acl",
						101,
						true,
						"6",
					),
					map[string]interface{}{
						"network_acl_id":  "acl",
						"egress":          true,
						"rule_action":     "allow",
						"cidr_block":      "172.31.0.0/16",
						"from_port":       80,
						"icmp_code":       0,
						"icmp_type":       0,
						"ipv6_cidr_block": "",
						"protocol":        "6",
						"rule_number":     101,
						"to_port":         80,
					},
				).Once().Return(&resource.Resource{
					Id:   "acl-rule2",
					Type: aws.AwsNetworkACLRuleResourceType,
				})

				factory.On(
					"CreateAbstractResource",
					aws.AwsNetworkACLRuleResourceType,
					aws.CreateNetworkACLRuleID(
						"acl",
						103,
						true,
						"6",
					),
					map[string]interface{}{
						"network_acl_id":  "acl",
						"egress":          true,
						"rule_action":     "deny",
						"cidr_block":      "172.31.0.0/16",
						"from_port":       80,
						"icmp_code":       0,
						"icmp_type":       0,
						"ipv6_cidr_block": "",
						"protocol":        "6",
						"rule_number":     103,
						"to_port":         80,
					},
				).Once().Return(&resource.Resource{
					Id:   "acl-rule3",
					Type: aws.AwsNetworkACLRuleResourceType,
				})

				factory.On(
					"CreateAbstractResource",
					aws.AwsNetworkACLRuleResourceType,
					aws.CreateNetworkACLRuleID(
						"default-acl",
						100,
						false,
						"17",
					),
					map[string]interface{}{
						"network_acl_id":  "default-acl",
						"egress":          false,
						"rule_action":     "allow",
						"cidr_block":      "172.31.0.0/16",
						"from_port":       80,
						"icmp_code":       0,
						"icmp_type":       0,
						"ipv6_cidr_block": "",
						"protocol":        "17",
						"rule_number":     100,
						"to_port":         80,
					},
				).Once().Return(&resource.Resource{
					Id:   "default-acl-rule1",
					Type: aws.AwsNetworkACLRuleResourceType,
				})
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:   "acl",
					Type: aws.AwsNetworkACLResourceType,
					Attrs: &resource.Attributes{
						"ingress": []interface{}{
							map[string]interface{}{
								"action":          "allow",
								"cidr_block":      "172.31.0.0/16",
								"from_port":       80,
								"icmp_code":       0,
								"icmp_type":       0,
								"ipv6_cidr_block": "",
								"protocol":        "17",
								"rule_no":         100,
								"to_port":         80,
							},
						},
						"egress": []interface{}{
							map[string]interface{}{
								"action":          "allow",
								"cidr_block":      "172.31.0.0/16",
								"from_port":       80,
								"icmp_code":       0,
								"icmp_type":       0,
								"ipv6_cidr_block": "",
								"protocol":        "6",
								"rule_no":         101,
								"to_port":         80,
							},
							// This one exist in state, test that we do not duplicate it
							// We map this expand to rule3 ID
							map[string]interface{}{
								"action":          "deny",
								"cidr_block":      "172.31.0.0/16",
								"from_port":       80,
								"icmp_code":       0,
								"icmp_type":       0,
								"ipv6_cidr_block": "",
								"protocol":        "6",
								"rule_no":         103,
								"to_port":         80,
							},
						},
					},
				},
				{
					Id:   "default-acl",
					Type: aws.AwsDefaultNetworkACLResourceType,
					Attrs: &resource.Attributes{
						"ingress": []interface{}{
							map[string]interface{}{
								"action":          "allow",
								"cidr_block":      "172.31.0.0/16",
								"from_port":       80,
								"icmp_code":       0,
								"icmp_type":       0,
								"ipv6_cidr_block": "",
								"protocol":        "17",
								"rule_no":         100,
								"to_port":         80,
							},
						},
					},
				},
				{
					Id:   "acl-rule3",
					Type: aws.AwsNetworkACLRuleResourceType,
				},
			},
			expectedFromRemote: []*resource.Resource{
				{
					Id: "fake",
				},
				{
					Id:    "ingress-and-egress-should-be-removed-from-remote-res",
					Type:  aws.AwsNetworkACLResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			expectedFromState: []*resource.Resource{
				{
					Id:    "acl",
					Type:  aws.AwsNetworkACLResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "acl-rule1",
					Type: aws.AwsNetworkACLRuleResourceType,
				},
				{
					Id:   "acl-rule2",
					Type: aws.AwsNetworkACLRuleResourceType,
				},
				{
					Id:   "acl-rule3",
					Type: aws.AwsNetworkACLRuleResourceType,
				},
				{
					Id:   "default-acl-rule1",
					Type: aws.AwsNetworkACLRuleResourceType,
				},
				{
					Id:    "default-acl",
					Type:  aws.AwsDefaultNetworkACLResourceType,
					Attrs: &resource.Attributes{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := &terraform.MockResourceFactory{}
			if tt.mock != nil {
				tt.mock(factory)
			}
			m := NewAwsNetworkACLExpander(factory)
			err := m.Execute(&tt.remoteResources, &tt.resourcesFromState)
			if err != nil {
				t.Fatal(err)
			}

			changelog, err := diff.Diff(tt.expectedFromRemote, tt.remoteResources)
			if err != nil {
				t.Fatal(err)
			}
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("expectedFromRemote %s got = %v, want %v", strings.Join(change.Path, "."), awsutil.Prettify(change.From), awsutil.Prettify(change.To))
				}
			}

			changelog, err = diff.Diff(tt.expectedFromState, tt.resourcesFromState)
			if err != nil {
				t.Fatal(err)
			}
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("expectedFromState %s got = %v, want %v", strings.Join(change.Path, "."), awsutil.Prettify(change.From), awsutil.Prettify(change.To))
				}
			}
		})
	}
}
