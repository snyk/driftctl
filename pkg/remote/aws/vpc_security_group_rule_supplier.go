package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/flatmap"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	sgRuleTypeIngress = "ingress"
	sgRuleTypeEgress  = "egress"
)

type VPCSecurityGroupRuleSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	repo         repository.EC2Repository
	runner       *terraform.ParallelResourceReader
}

type securityGroupRule struct {
	Type                  string
	SecurityGroupId       string
	Protocol              string
	FromPort              int
	ToPort                int
	Self                  bool
	SourceSecurityGroupId string
	CidrBlocks            []string
	Ipv6CidrBlocks        []string
	PrefixListIds         []string
}

func (s *securityGroupRule) getId() string {
	attrs := resource.Attributes{
		"type":                     s.Type,
		"security_group_id":        s.SecurityGroupId,
		"protocol":                 s.Protocol,
		"from_port":                float64(s.FromPort),
		"to_port":                  float64(s.ToPort),
		"self":                     s.Self,
		"source_security_group_id": s.SourceSecurityGroupId,
		"cidr_blocks":              toInterfaceSlice(s.CidrBlocks),
		"ipv6_cidr_blocks":         toInterfaceSlice(s.Ipv6CidrBlocks),
		"prefix_list_ids":          toInterfaceSlice(s.PrefixListIds),
	}

	return resourceaws.CreateSecurityGroupRuleIdHash(&attrs)
}

func toInterfaceSlice(val []string) []interface{} {
	var res []interface{}
	for _, v := range val {
		res = append(res, v)
	}
	return res
}

func NewVPCSecurityGroupRuleSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.EC2Repository) *VPCSecurityGroupRuleSupplier {
	return &VPCSecurityGroupRuleSupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *VPCSecurityGroupRuleSupplier) SuppliedType() resource.ResourceType {
	return resourceaws.AwsSecurityGroupRuleResourceType
}

func (s *VPCSecurityGroupRuleSupplier) Resources() ([]resource.Resource, error) {
	securityGroups, defaultSecurityGroups, err := s.repo.ListAllSecurityGroups()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
	}
	secGroups := make([]*ec2.SecurityGroup, 0, len(securityGroups)+len(defaultSecurityGroups))
	secGroups = append(secGroups, securityGroups...)
	secGroups = append(secGroups, defaultSecurityGroups...)
	securityGroupsRules := s.listSecurityGroupsRules(secGroups)
	results := make([]cty.Value, 0)
	if len(securityGroupsRules) > 0 {
		for _, securityGroupsRule := range securityGroupsRules {
			rule := securityGroupsRule
			s.runner.Run(func() (cty.Value, error) {
				return s.readSecurityGroupRule(rule)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(s.SuppliedType(), results)
}

func (s *VPCSecurityGroupRuleSupplier) readSecurityGroupRule(rule securityGroupRule) (cty.Value, error) {
	id := rule.getId()

	resSgRule, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: s.SuppliedType(),
		ID: id,
		Attributes: flatmap.Flatten(map[string]interface{}{
			"type":                     rule.Type,
			"security_group_id":        rule.SecurityGroupId,
			"protocol":                 rule.Protocol,
			"from_port":                rule.FromPort,
			"to_port":                  rule.ToPort,
			"self":                     rule.Self,
			"source_security_group_id": rule.SourceSecurityGroupId,
			"cidr_blocks":              rule.CidrBlocks,
			"ipv6_cidr_blocks":         rule.Ipv6CidrBlocks,
			"prefix_list_ids":          rule.PrefixListIds,
		}),
	})
	if err != nil {
		logrus.Warnf("Error reading rule from security group %s: %+v", id, err)
		return cty.NilVal, err
	}
	return *resSgRule, nil
}

func (s *VPCSecurityGroupRuleSupplier) listSecurityGroupsRules(securityGroups []*ec2.SecurityGroup) []securityGroupRule {
	var securityGroupsRules []securityGroupRule
	for _, sg := range securityGroups {
		for _, rule := range sg.IpPermissions {
			securityGroupsRules = append(securityGroupsRules, s.addSecurityGroupRule(sgRuleTypeIngress, rule, sg)...)
		}
		for _, rule := range sg.IpPermissionsEgress {
			securityGroupsRules = append(securityGroupsRules, s.addSecurityGroupRule(sgRuleTypeEgress, rule, sg)...)
		}
	}
	return securityGroupsRules
}

// addSecurityGroupRule will iterate through each "Source" as per Aws definition and create a
// rule with custom attributes
func (s *VPCSecurityGroupRuleSupplier) addSecurityGroupRule(ruleType string, rule *ec2.IpPermission, sg *ec2.SecurityGroup) []securityGroupRule {
	var rules []securityGroupRule
	for _, groupPair := range rule.UserIdGroupPairs {
		r := securityGroupRule{
			Type:            ruleType,
			SecurityGroupId: aws.StringValue(sg.GroupId),
			Protocol:        aws.StringValue(rule.IpProtocol),
			FromPort:        int(aws.Int64Value(rule.FromPort)),
			ToPort:          int(aws.Int64Value(rule.ToPort)),
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
			FromPort:        int(aws.Int64Value(rule.FromPort)),
			ToPort:          int(aws.Int64Value(rule.ToPort)),
			CidrBlocks:      []string{aws.StringValue(ipRange.CidrIp)},
		}
		rules = append(rules, r)
	}
	for _, ipRange := range rule.Ipv6Ranges {
		r := securityGroupRule{
			Type:            ruleType,
			SecurityGroupId: aws.StringValue(sg.GroupId),
			Protocol:        aws.StringValue(rule.IpProtocol),
			FromPort:        int(aws.Int64Value(rule.FromPort)),
			ToPort:          int(aws.Int64Value(rule.ToPort)),
			Ipv6CidrBlocks:  []string{aws.StringValue(ipRange.CidrIpv6)},
		}
		rules = append(rules, r)
	}
	for _, listId := range rule.PrefixListIds {
		r := securityGroupRule{
			Type:            ruleType,
			SecurityGroupId: aws.StringValue(sg.GroupId),
			Protocol:        aws.StringValue(rule.IpProtocol),
			FromPort:        int(aws.Int64Value(rule.FromPort)),
			ToPort:          int(aws.Int64Value(rule.ToPort)),
			PrefixListIds:   []string{aws.StringValue(listId.PrefixListId)},
		}
		rules = append(rules, r)
	}
	// Filter default rules for default security group
	if sg.GroupName != nil && *sg.GroupName == "default" {
		results := make([]securityGroupRule, 0, len(rules))
		for _, r := range rules {
			r := r
			if s.isDefaultIngress(&r) || s.isDefaultEgress(&r) {
				continue
			}
			results = append(results, r)
		}
		return results
	}

	return rules
}

func (s *VPCSecurityGroupRuleSupplier) isDefaultIngress(rule *securityGroupRule) bool {
	return rule.Type == sgRuleTypeIngress &&
		rule.FromPort == 0 &&
		rule.ToPort == 0 &&
		rule.Protocol == "-1" &&
		rule.CidrBlocks == nil &&
		rule.Ipv6CidrBlocks == nil &&
		rule.PrefixListIds == nil &&
		rule.SourceSecurityGroupId == "" &&
		rule.Self
}

func (s *VPCSecurityGroupRuleSupplier) isDefaultEgress(rule *securityGroupRule) bool {
	return rule.Type == sgRuleTypeEgress &&
		rule.FromPort == 0 &&
		rule.ToPort == 0 &&
		rule.Protocol == "-1" &&
		rule.Ipv6CidrBlocks == nil &&
		rule.PrefixListIds == nil &&
		rule.SourceSecurityGroupId == "" &&
		rule.CidrBlocks != nil &&
		len(rule.CidrBlocks) == 1 &&
		(rule.CidrBlocks)[0] == "0.0.0.0/0"
}
