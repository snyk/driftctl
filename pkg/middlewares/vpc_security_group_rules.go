package middlewares

import (
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
)

// Split security group rule if it needs to given its attributes
type VPCSecurityGroupRuleSanitizer struct {
	resourceFactory resource.ResourceFactory
}

func NewVPCSecurityGroupRuleSanitizer(resourceFactory resource.ResourceFactory) VPCSecurityGroupRuleSanitizer {
	return VPCSecurityGroupRuleSanitizer{
		resourceFactory,
	}
}

func (m VPCSecurityGroupRuleSanitizer) Execute(_, resourcesFromState *[]resource.Resource) error {
	newStateResources := make([]resource.Resource, 0)

	for _, stateResource := range *resourcesFromState {
		// Ignore all resources other than security group rule
		if stateResource.TerraformType() != resourceaws.AwsSecurityGroupRuleResourceType {
			newStateResources = append(newStateResources, stateResource)
			continue
		}

		securityGroupRule, _ := stateResource.(*resourceaws.AwsSecurityGroupRule)

		if split := shouldBeSplit(securityGroupRule); !split {
			newStateResources = append(newStateResources, stateResource)
			continue
		}

		if securityGroupRule.CidrBlocks != nil && len(*securityGroupRule.CidrBlocks) > 0 {
			for _, ipRange := range *securityGroupRule.CidrBlocks {
				rule := *securityGroupRule
				rule.CidrBlocks = &[]string{ipRange}
				rule.Ipv6CidrBlocks = &[]string{}
				rule.PrefixListIds = &[]string{}
				res, err := m.createRule(&rule)
				if err != nil {
					return err
				}
				logrus.WithFields(logrus.Fields{
					"formerRuleId": securityGroupRule.TerraformId(),
					"newRuleId":    rule.TerraformId(),
				}).Debug("Splitting aws_security_group_rule")
				newStateResources = append(newStateResources, res)
			}
		}
		if securityGroupRule.Ipv6CidrBlocks != nil && len(*securityGroupRule.Ipv6CidrBlocks) > 0 {
			for _, ipRange := range *securityGroupRule.Ipv6CidrBlocks {
				rule := *securityGroupRule
				rule.CidrBlocks = &[]string{}
				rule.Ipv6CidrBlocks = &[]string{ipRange}
				rule.PrefixListIds = &[]string{}
				res, err := m.createRule(&rule)
				if err != nil {
					return err
				}
				logrus.WithFields(logrus.Fields{
					"formerRuleId": securityGroupRule.TerraformId(),
					"newRuleId":    rule.TerraformId(),
				}).Debug("Splitting aws_security_group_rule")
				newStateResources = append(newStateResources, res)
			}
		}
		if securityGroupRule.PrefixListIds != nil && len(*securityGroupRule.PrefixListIds) > 0 {
			for _, listId := range *securityGroupRule.PrefixListIds {
				rule := *securityGroupRule
				rule.CidrBlocks = &[]string{}
				rule.Ipv6CidrBlocks = &[]string{}
				rule.PrefixListIds = &[]string{listId}
				res, err := m.createRule(&rule)
				if err != nil {
					return err
				}
				logrus.WithFields(logrus.Fields{
					"formerRuleId": securityGroupRule.TerraformId(),
					"newRuleId":    rule.TerraformId(),
				}).Debug("Splitting aws_security_group_rule")
				newStateResources = append(newStateResources, res)
			}
		}
		if (securityGroupRule.Self != nil && *securityGroupRule.Self) ||
			(securityGroupRule.SourceSecurityGroupId != nil && *securityGroupRule.SourceSecurityGroupId != "") {
			rule := *securityGroupRule
			rule.CidrBlocks = &[]string{}
			rule.Ipv6CidrBlocks = &[]string{}
			rule.PrefixListIds = &[]string{}
			res, err := m.createRule(&rule)
			if err != nil {
				return err
			}
			rule.Id = rule.CreateIdHash()
			logrus.WithFields(logrus.Fields{
				"formerRuleId": securityGroupRule.TerraformId(),
				"newRuleId":    rule.TerraformId(),
			}).Debug("Splitting aws_security_group_rule")
			newStateResources = append(newStateResources, res)
		}
	}

	*resourcesFromState = newStateResources

	return nil
}

func (m *VPCSecurityGroupRuleSanitizer) createRule(res *resourceaws.AwsSecurityGroupRule) (*resourceaws.AwsSecurityGroupRule, error) {
	res.Id = res.CreateIdHash()
	data := map[string]interface{}{
		"id":                       res.Id,
		"cidr_blocks":              res.CidrBlocks,
		"description":              res.Description,
		"from_port":                res.FromPort,
		"ipv6_cidr_blocks":         res.Ipv6CidrBlocks,
		"prefix_list_ids":          res.PrefixListIds,
		"protocol":                 res.Protocol,
		"security_group_id":        res.SecurityGroupId,
		"self":                     res.Self,
		"source_security_group_id": res.SourceSecurityGroupId,
		"to_port":                  res.ToPort,
		"type":                     res.Type,
	}
	ctyVal, err := m.resourceFactory.CreateResource(data, "aws_security_group_rule")
	if err != nil {
		return nil, err
	}
	res.CtyVal = ctyVal
	return res, err
}

func shouldBeSplit(rule *resourceaws.AwsSecurityGroupRule) bool {
	var i int
	if rule.CidrBlocks != nil && len(*rule.CidrBlocks) > 0 {
		i += len(*rule.CidrBlocks)
	}
	if rule.Ipv6CidrBlocks != nil && len(*rule.Ipv6CidrBlocks) > 0 {
		i += len(*rule.Ipv6CidrBlocks)
	}
	if rule.PrefixListIds != nil && len(*rule.PrefixListIds) > 0 {
		i += len(*rule.PrefixListIds)
	}
	if (rule.Self != nil && *rule.Self) ||
		(rule.SourceSecurityGroupId != nil && *rule.SourceSecurityGroupId != "") {
		i += 1
	}
	return i > 1
}
