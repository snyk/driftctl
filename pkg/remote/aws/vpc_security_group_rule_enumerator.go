package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"

	"github.com/snyk/driftctl/pkg/resource"
	resourceaws "github.com/snyk/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const (
	sgRuleTypeIngress = "ingress"
	sgRuleTypeEgress  = "egress"
)

type VPCSecurityGroupRuleEnumerator struct {
	repository repository.EC2Repository
	factory    resource.ResourceFactory
}

type securityGroupRule struct {
	Type                  string
	SecurityGroupId       string
	Protocol              string
	FromPort              float64
	ToPort                float64
	Self                  bool
	SourceSecurityGroupId string
	CidrBlocks            []string
	Ipv6CidrBlocks        []string
	PrefixListIds         []string
}

func (s *securityGroupRule) getId() string {
	attrs := s.getAttrs()
	return resourceaws.CreateSecurityGroupRuleIdHash(&attrs)
}

func (s *securityGroupRule) getAttrs() resource.Attributes {
	attrs := resource.Attributes{
		"type":                     s.Type,
		"security_group_id":        s.SecurityGroupId,
		"protocol":                 s.Protocol,
		"from_port":                s.FromPort,
		"to_port":                  s.ToPort,
		"self":                     s.Self,
		"source_security_group_id": s.SourceSecurityGroupId,
		"cidr_blocks":              toInterfaceSlice(s.CidrBlocks),
		"ipv6_cidr_blocks":         toInterfaceSlice(s.Ipv6CidrBlocks),
		"prefix_list_ids":          toInterfaceSlice(s.PrefixListIds),
	}

	return attrs
}

func toInterfaceSlice(val []string) []interface{} {
	var res []interface{}
	for _, v := range val {
		res = append(res, v)
	}
	return res
}

func NewVPCSecurityGroupRuleEnumerator(repository repository.EC2Repository, factory resource.ResourceFactory) *VPCSecurityGroupRuleEnumerator {
	return &VPCSecurityGroupRuleEnumerator{
		repository,
		factory,
	}
}

func (e *VPCSecurityGroupRuleEnumerator) SupportedType() resource.ResourceType {
	return resourceaws.AwsSecurityGroupRuleResourceType
}

func (e *VPCSecurityGroupRuleEnumerator) Enumerate() ([]*resource.Resource, error) {
	securityGroups, defaultSecurityGroups, err := e.repository.ListAllSecurityGroups()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), resourceaws.AwsSecurityGroupResourceType)
	}

	secGroups := make([]*ec2.SecurityGroup, 0, len(securityGroups)+len(defaultSecurityGroups))
	secGroups = append(secGroups, securityGroups...)
	secGroups = append(secGroups, defaultSecurityGroups...)
	securityGroupsRules := e.listSecurityGroupsRules(secGroups)

	results := make([]*resource.Resource, 0, len(securityGroupsRules))
	for _, rule := range securityGroupsRules {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				rule.getId(),
				rule.getAttrs(),
			),
		)
	}

	return results, nil
}

func (e *VPCSecurityGroupRuleEnumerator) listSecurityGroupsRules(securityGroups []*ec2.SecurityGroup) []securityGroupRule {
	var securityGroupsRules []securityGroupRule
	for _, sg := range securityGroups {
		for _, rule := range sg.IpPermissions {
			securityGroupsRules = append(securityGroupsRules, e.addSecurityGroupRule(sgRuleTypeIngress, rule, sg)...)
		}
		for _, rule := range sg.IpPermissionsEgress {
			securityGroupsRules = append(securityGroupsRules, e.addSecurityGroupRule(sgRuleTypeEgress, rule, sg)...)
		}
	}
	return securityGroupsRules
}

// addSecurityGroupRule will iterate through each "Source" as per Aws definition and create a
// rule with custom attributes
func (e *VPCSecurityGroupRuleEnumerator) addSecurityGroupRule(ruleType string, rule *ec2.IpPermission, sg *ec2.SecurityGroup) []securityGroupRule {
	var rules []securityGroupRule
	for _, groupPair := range rule.UserIdGroupPairs {
		r := securityGroupRule{
			Type:            ruleType,
			SecurityGroupId: aws.StringValue(sg.GroupId),
			Protocol:        aws.StringValue(rule.IpProtocol),
			FromPort:        float64(aws.Int64Value(rule.FromPort)),
			ToPort:          float64(aws.Int64Value(rule.ToPort)),
		}
		if aws.StringValue(groupPair.GroupId) == aws.StringValue(sg.GroupId) {
			r.Self = true
		} else {
			r.SourceSecurityGroupId = aws.StringValue(groupPair.GroupId)
		}
		rules = append(rules, r)
	}
	for _, ipRange := range rule.IpRanges {
		r := securityGroupRule{
			Type:            ruleType,
			SecurityGroupId: aws.StringValue(sg.GroupId),
			Protocol:        aws.StringValue(rule.IpProtocol),
			FromPort:        float64(aws.Int64Value(rule.FromPort)),
			ToPort:          float64(aws.Int64Value(rule.ToPort)),
			CidrBlocks:      []string{aws.StringValue(ipRange.CidrIp)},
		}
		rules = append(rules, r)
	}
	for _, ipRange := range rule.Ipv6Ranges {
		r := securityGroupRule{
			Type:            ruleType,
			SecurityGroupId: aws.StringValue(sg.GroupId),
			Protocol:        aws.StringValue(rule.IpProtocol),
			FromPort:        float64(aws.Int64Value(rule.FromPort)),
			ToPort:          float64(aws.Int64Value(rule.ToPort)),
			Ipv6CidrBlocks:  []string{aws.StringValue(ipRange.CidrIpv6)},
		}
		rules = append(rules, r)
	}
	for _, listId := range rule.PrefixListIds {
		r := securityGroupRule{
			Type:            ruleType,
			SecurityGroupId: aws.StringValue(sg.GroupId),
			Protocol:        aws.StringValue(rule.IpProtocol),
			FromPort:        float64(aws.Int64Value(rule.FromPort)),
			ToPort:          float64(aws.Int64Value(rule.ToPort)),
			PrefixListIds:   []string{aws.StringValue(listId.PrefixListId)},
		}
		rules = append(rules, r)
	}
	return rules
}
