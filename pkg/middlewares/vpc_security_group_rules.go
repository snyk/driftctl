package middlewares

import (
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
)

// Split security group rule if it needs to given its attributes
type VPCSecurityGroupRuleSanitizer struct{}

func NewVPCSecurityGroupRuleSanitizer() VPCSecurityGroupRuleSanitizer {
	return VPCSecurityGroupRuleSanitizer{}
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
				rule := resourceaws.AwsSecurityGroupRule{
					Type:            securityGroupRule.Type,
					Description:     securityGroupRule.Description,
					SecurityGroupId: securityGroupRule.SecurityGroupId,
					Protocol:        securityGroupRule.Protocol,
					FromPort:        securityGroupRule.FromPort,
					ToPort:          securityGroupRule.ToPort,
					CidrBlocks:      &[]string{ipRange},
					Ipv6CidrBlocks:  &[]string{},
					PrefixListIds:   &[]string{},
				}
				rule.Id = rule.CreateIdHash()
				logrus.WithFields(logrus.Fields{
					"formerRuleId": securityGroupRule.TerraformId(),
					"newRuleId":    rule.TerraformId(),
				}).Debug("Splitting aws_security_group_rule")
				newStateResources = append(newStateResources, &rule)
			}
		}
		if securityGroupRule.Ipv6CidrBlocks != nil && len(*securityGroupRule.Ipv6CidrBlocks) > 0 {
			for _, ipRange := range *securityGroupRule.Ipv6CidrBlocks {
				rule := resourceaws.AwsSecurityGroupRule{
					Type:            securityGroupRule.Type,
					Description:     securityGroupRule.Description,
					SecurityGroupId: securityGroupRule.SecurityGroupId,
					Protocol:        securityGroupRule.Protocol,
					FromPort:        securityGroupRule.FromPort,
					ToPort:          securityGroupRule.ToPort,
					CidrBlocks:      &[]string{},
					Ipv6CidrBlocks:  &[]string{ipRange},
					PrefixListIds:   &[]string{},
				}
				rule.Id = rule.CreateIdHash()
				logrus.WithFields(logrus.Fields{
					"formerRuleId": securityGroupRule.TerraformId(),
					"newRuleId":    rule.TerraformId(),
				}).Debug("Splitting aws_security_group_rule")
				newStateResources = append(newStateResources, &rule)
			}
		}
		if securityGroupRule.PrefixListIds != nil && len(*securityGroupRule.PrefixListIds) > 0 {
			for _, listId := range *securityGroupRule.PrefixListIds {
				rule := resourceaws.AwsSecurityGroupRule{
					Type:            securityGroupRule.Type,
					Description:     securityGroupRule.Description,
					SecurityGroupId: securityGroupRule.SecurityGroupId,
					Protocol:        securityGroupRule.Protocol,
					FromPort:        securityGroupRule.FromPort,
					ToPort:          securityGroupRule.ToPort,
					CidrBlocks:      &[]string{},
					Ipv6CidrBlocks:  &[]string{},
					PrefixListIds:   &[]string{listId},
				}
				rule.Id = rule.CreateIdHash()
				logrus.WithFields(logrus.Fields{
					"formerRuleId": securityGroupRule.TerraformId(),
					"newRuleId":    rule.TerraformId(),
				}).Debug("Splitting aws_security_group_rule")
				newStateResources = append(newStateResources, &rule)
			}
		}
		if (securityGroupRule.Self != nil && *securityGroupRule.Self) ||
			(securityGroupRule.SourceSecurityGroupId != nil && *securityGroupRule.SourceSecurityGroupId != "") {
			rule := resourceaws.AwsSecurityGroupRule{
				Type:                  securityGroupRule.Type,
				Description:           securityGroupRule.Description,
				SecurityGroupId:       securityGroupRule.SecurityGroupId,
				Protocol:              securityGroupRule.Protocol,
				FromPort:              securityGroupRule.FromPort,
				ToPort:                securityGroupRule.ToPort,
				CidrBlocks:            &[]string{},
				Ipv6CidrBlocks:        &[]string{},
				PrefixListIds:         &[]string{},
				Self:                  securityGroupRule.Self,
				SourceSecurityGroupId: securityGroupRule.SourceSecurityGroupId,
			}
			rule.Id = rule.CreateIdHash()
			logrus.WithFields(logrus.Fields{
				"formerRuleId": securityGroupRule.TerraformId(),
				"newRuleId":    rule.TerraformId(),
			}).Debug("Splitting aws_security_group_rule")
			newStateResources = append(newStateResources, &rule)
		}
	}

	*resourcesFromState = newStateResources

	return nil
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
