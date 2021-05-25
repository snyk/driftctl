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

func (m VPCSecurityGroupRuleSanitizer) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {
	newStateResources := make([]resource.Resource, 0)

	for _, stateResource := range *resourcesFromState {
		// Ignore all resources other than security group rule
		if stateResource.TerraformType() != resourceaws.AwsSecurityGroupRuleResourceType {
			newStateResources = append(newStateResources, stateResource)
			continue
		}

		rule, _ := stateResource.(*resource.AbstractResource)

		if rule.Attrs.GetBool("self") != nil && *rule.Attrs.GetBool("self") {
			_ = rule.Attrs.SafeSet([]string{"source_security_group_id"}, *rule.Attrs.GetString("security_group_id"))
		}

		if !shouldBeSplit(rule) {
			rule.Attrs.SafeDelete([]string{"self"})
			newStateResources = append(newStateResources, stateResource)
			continue
		}

		if rule.Attrs.GetSlice("cidr_blocks") != nil && len(rule.Attrs.GetSlice("cidr_blocks")) > 0 {
			for _, ipRange := range rule.Attrs.GetSlice("cidr_blocks") {
				attrs := rule.Attrs.Copy()
				_ = attrs.SafeSet([]string{"cidr_blocks"}, []interface{}{ipRange})
				_ = attrs.SafeSet([]string{"ipv6_cidr_blocks"}, []interface{}{})
				_ = attrs.SafeSet([]string{"prefix_list_ids"}, []interface{}{})
				res := m.createRule(attrs)
				logrus.WithFields(logrus.Fields{
					"formerRuleId": rule.TerraformId(),
					"newRuleId":    res.TerraformId(),
				}).Debug("Splitting aws_security_group_rule")
				res.Attrs.SafeDelete([]string{"self"})
				newStateResources = append(newStateResources, res)
			}
		}

		if rule.Attrs.GetSlice("ipv6_cidr_blocks") != nil && len(rule.Attrs.GetSlice("ipv6_cidr_blocks")) > 0 {
			for _, ipRange := range rule.Attrs.GetSlice("ipv6_cidr_blocks") {
				attrs := rule.Attrs.Copy()
				_ = attrs.SafeSet([]string{"cidr_blocks"}, []interface{}{})
				_ = attrs.SafeSet([]string{"ipv6_cidr_blocks"}, []interface{}{ipRange})
				_ = attrs.SafeSet([]string{"prefix_list_ids"}, []interface{}{})
				res := m.createRule(attrs)
				logrus.WithFields(logrus.Fields{
					"formerRuleId": rule.TerraformId(),
					"newRuleId":    res.TerraformId(),
				}).Debug("Splitting aws_security_group_rule")
				res.Attrs.SafeDelete([]string{"self"})
				newStateResources = append(newStateResources, res)
			}
		}

		if rule.Attrs.GetSlice("prefix_list_ids") != nil && len(rule.Attrs.GetSlice("prefix_list_ids")) > 0 {
			for _, listId := range rule.Attrs.GetSlice("prefix_list_ids") {
				attrs := rule.Attrs.Copy()
				_ = attrs.SafeSet([]string{"cidr_blocks"}, []interface{}{})
				_ = attrs.SafeSet([]string{"ipv6_cidr_blocks"}, []interface{}{})
				_ = attrs.SafeSet([]string{"prefix_list_ids"}, []interface{}{listId})
				res := m.createRule(attrs)
				logrus.WithFields(logrus.Fields{
					"formerRuleId": rule.TerraformId(),
					"newRuleId":    res.TerraformId(),
				}).Debug("Splitting aws_security_group_rule")
				res.Attrs.SafeDelete([]string{"self"})
				newStateResources = append(newStateResources, res)
			}
		}

		if (rule.Attrs.GetBool("self") != nil && *rule.Attrs.GetBool("self")) ||
			(rule.Attrs.GetString("source_security_group_id") != nil && *rule.Attrs.GetString("source_security_group_id") != "") {
			attrs := rule.Attrs.Copy()
			_ = attrs.SafeSet([]string{"cidr_blocks"}, []interface{}{})
			_ = attrs.SafeSet([]string{"ipv6_cidr_blocks"}, []interface{}{})
			_ = attrs.SafeSet([]string{"prefix_list_ids"}, []interface{}{})
			res := m.createRule(attrs)
			logrus.WithFields(logrus.Fields{
				"formerRuleId": rule.TerraformId(),
				"newRuleId":    res.TerraformId(),
			}).Debug("Splitting aws_security_group_rule")
			res.Attrs.SafeDelete([]string{"self"})
			newStateResources = append(newStateResources, res)
		}
	}

	*resourcesFromState = newStateResources

	for _, res := range *remoteResources {
		if res.TerraformType() != resourceaws.AwsSecurityGroupRuleResourceType {
			continue
		}
		rule, _ := res.(*resource.AbstractResource)
		rule.Attrs.SafeDelete([]string{"self"})
	}

	return nil
}

func (m *VPCSecurityGroupRuleSanitizer) createRule(res *resource.Attributes) *resource.AbstractResource {
	id := resourceaws.CreateSecurityGroupRuleIdHash(res)
	data := map[string]interface{}{
		"id":                       id,
		"cidr_blocks":              (*res)["cidr_blocks"],
		"description":              (*res)["description"],
		"from_port":                (*res)["from_port"],
		"ipv6_cidr_blocks":         (*res)["ipv6_cidr_blocks"],
		"prefix_list_ids":          (*res)["prefix_list_ids"],
		"protocol":                 (*res)["protocol"],
		"security_group_id":        (*res)["security_group_id"],
		"self":                     (*res)["self"],
		"source_security_group_id": (*res)["source_security_group_id"],
		"to_port":                  (*res)["to_port"],
		"type":                     (*res)["type"],
	}
	rule := m.resourceFactory.CreateAbstractResource("aws_security_group_rule", id, data)
	return rule
}

func shouldBeSplit(r *resource.AbstractResource) bool {
	var i int
	if r.Attrs.GetSlice("cidr_blocks") != nil && len(r.Attrs.GetSlice("cidr_blocks")) > 0 {
		i += len(r.Attrs.GetSlice("cidr_blocks"))
	}

	if r.Attrs.GetSlice("ipv6_cidr_blocks") != nil && len(r.Attrs.GetSlice("ipv6_cidr_blocks")) > 0 {
		i += len(r.Attrs.GetSlice("ipv6_cidr_blocks"))
	}

	if r.Attrs.GetSlice("prefix_list_ids") != nil && len(r.Attrs.GetSlice("prefix_list_ids")) > 0 {
		i += len(r.Attrs.GetSlice("prefix_list_ids"))
	}

	if r.Attrs.GetBool("self") != nil && *r.Attrs.GetBool("self") ||
		(r.Attrs.GetString("source_security_group_id") != nil && *r.Attrs.GetString("source_security_group_id") != "") {
		i += 1
	}
	return i > 1
}
