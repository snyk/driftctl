package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
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
	deserializer deserializer.CTYDeserializer
	client       repository.EC2Repository
	runner       *terraform.ParallelResourceReader
}

func NewVPCSecurityGroupRuleSupplier(provider *AWSTerraformProvider) *VPCSecurityGroupRuleSupplier {
	return &VPCSecurityGroupRuleSupplier{
		provider,
		awsdeserializer.NewVPCSecurityGroupRuleDeserializer(),
		repository.NewEC2Repository(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *VPCSecurityGroupRuleSupplier) Resources() ([]resource.Resource, error) {
	securityGroups, defaultSecurityGroups, err := s.client.ListAllSecurityGroups()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsSecurityGroupRuleResourceType)
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
	return s.deserializer.Deserialize(results)
}

func (s *VPCSecurityGroupRuleSupplier) readSecurityGroupRule(securityGroupRule resourceaws.AwsSecurityGroupRule) (cty.Value, error) {
	id := securityGroupRule.Id
	f := func(v *[]string) []string {
		if v != nil {
			return *v
		}
		return []string{}
	}
	resSgRule, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: resourceaws.AwsSecurityGroupRuleResourceType,
		ID: id,
		Attributes: flatmap.Flatten(map[string]interface{}{
			"type":                     aws.StringValue(securityGroupRule.Type),
			"security_group_id":        aws.StringValue(securityGroupRule.SecurityGroupId),
			"protocol":                 aws.StringValue(securityGroupRule.Protocol),
			"from_port":                aws.IntValue(securityGroupRule.FromPort),
			"to_port":                  aws.IntValue(securityGroupRule.ToPort),
			"self":                     aws.BoolValue(securityGroupRule.Self),
			"source_security_group_id": aws.StringValue(securityGroupRule.SourceSecurityGroupId),
			"cidr_blocks":              f(securityGroupRule.CidrBlocks),
			"ipv6_cidr_blocks":         f(securityGroupRule.Ipv6CidrBlocks),
			"prefix_list_ids":          f(securityGroupRule.PrefixListIds),
		}),
	})
	if err != nil {
		logrus.Warnf("Error reading rule from security group %s: %+v", id, err)
		return cty.NilVal, err
	}
	return *resSgRule, nil
}

func (s *VPCSecurityGroupRuleSupplier) listSecurityGroupsRules(securityGroups []*ec2.SecurityGroup) []resourceaws.AwsSecurityGroupRule {
	var securityGroupsRules []resourceaws.AwsSecurityGroupRule
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
func (s *VPCSecurityGroupRuleSupplier) addSecurityGroupRule(ruleType string, rule *ec2.IpPermission, sg *ec2.SecurityGroup) []resourceaws.AwsSecurityGroupRule {
	var rules []resourceaws.AwsSecurityGroupRule
	for _, groupPair := range rule.UserIdGroupPairs {
		r := resourceaws.AwsSecurityGroupRule{
			Type:            aws.String(ruleType),
			SecurityGroupId: sg.GroupId,
			Protocol:        rule.IpProtocol,
			FromPort:        aws.Int(int(aws.Int64Value(rule.FromPort))),
			ToPort:          aws.Int(int(aws.Int64Value(rule.ToPort))),
		}
		if aws.StringValue(groupPair.GroupId) == aws.StringValue(sg.GroupId) {
			r.Self = aws.Bool(true)
		} else {
			r.SourceSecurityGroupId = groupPair.GroupId
		}
		r.Id = r.CreateIdHash()
		rules = append(rules, r)
	}
	for _, ipRange := range rule.IpRanges {
		r := resourceaws.AwsSecurityGroupRule{
			Type:            aws.String(ruleType),
			SecurityGroupId: sg.GroupId,
			Protocol:        rule.IpProtocol,
			FromPort:        aws.Int(int(aws.Int64Value(rule.FromPort))),
			ToPort:          aws.Int(int(aws.Int64Value(rule.ToPort))),
			CidrBlocks:      &[]string{aws.StringValue(ipRange.CidrIp)},
		}
		r.Id = r.CreateIdHash()
		rules = append(rules, r)
	}
	for _, ipRange := range rule.Ipv6Ranges {
		r := resourceaws.AwsSecurityGroupRule{
			Type:            aws.String(ruleType),
			SecurityGroupId: sg.GroupId,
			Protocol:        rule.IpProtocol,
			FromPort:        aws.Int(int(aws.Int64Value(rule.FromPort))),
			ToPort:          aws.Int(int(aws.Int64Value(rule.ToPort))),
			Ipv6CidrBlocks:  &[]string{aws.StringValue(ipRange.CidrIpv6)},
		}
		r.Id = r.CreateIdHash()
		rules = append(rules, r)
	}
	for _, listId := range rule.PrefixListIds {
		r := resourceaws.AwsSecurityGroupRule{
			Type:            aws.String(ruleType),
			SecurityGroupId: sg.GroupId,
			Protocol:        rule.IpProtocol,
			FromPort:        aws.Int(int(aws.Int64Value(rule.FromPort))),
			ToPort:          aws.Int(int(aws.Int64Value(rule.ToPort))),
			PrefixListIds:   &[]string{aws.StringValue(listId.PrefixListId)},
		}
		r.Id = r.CreateIdHash()
		rules = append(rules, r)
	}
	// Filter default rules for default security group
	if sg.GroupName != nil && *sg.GroupName == "default" {
		results := make([]resourceaws.AwsSecurityGroupRule, 0, len(rules))
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

func (s *VPCSecurityGroupRuleSupplier) isDefaultIngress(rule *resourceaws.AwsSecurityGroupRule) bool {
	return rule.Type != nil &&
		*rule.Type == sgRuleTypeIngress &&
		rule.FromPort != nil &&
		*rule.FromPort == 0 &&
		rule.ToPort != nil &&
		*rule.ToPort == 0 &&
		rule.Protocol != nil &&
		*rule.Protocol == "-1" &&
		rule.CidrBlocks == nil &&
		rule.Ipv6CidrBlocks == nil &&
		rule.PrefixListIds == nil &&
		rule.SourceSecurityGroupId == nil &&
		rule.Self != nil &&
		*rule.Self
}

func (s *VPCSecurityGroupRuleSupplier) isDefaultEgress(rule *resourceaws.AwsSecurityGroupRule) bool {
	return rule.Type != nil &&
		*rule.Type == sgRuleTypeEgress &&
		rule.FromPort != nil &&
		*rule.FromPort == 0 &&
		rule.ToPort != nil &&
		*rule.ToPort == 0 &&
		rule.Protocol != nil &&
		*rule.Protocol == "-1" &&
		rule.Ipv6CidrBlocks == nil &&
		rule.PrefixListIds == nil &&
		rule.SourceSecurityGroupId == nil &&
		rule.CidrBlocks != nil &&
		len(*rule.CidrBlocks) == 1 &&
		(*rule.CidrBlocks)[0] == "0.0.0.0/0"
}
