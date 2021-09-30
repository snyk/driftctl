package middlewares

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

// Default network acl rules should not be shown as unmanaged as they are present by default
// This middleware ignores default network acl rules from unmanaged resources if they are not managed by IaC
type AwsDefaultNetworkACLRule struct{}

func NewAwsDefaultNetworkACLRule() AwsDefaultNetworkACLRule {
	return AwsDefaultNetworkACLRule{}
}

func (m AwsDefaultNetworkACLRule) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {

	newRemoteResources := make([]*resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than ACL rules
		if remoteResource.ResourceType() != aws.AwsNetworkACLRuleResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Ignore non default ACL rules
		if !m.isDefaultACLRule(remoteResource) {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Check if resource is managed by IaC
		existInState := false
		for _, stateResource := range *resourcesFromState {
			if remoteResource.Equal(stateResource) {
				existInState = true
				break
			}
		}

		// Include resource if it's managed in IaC
		if existInState {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Else, resource is not added to newRemoteResources slice so it will be ignored
		logrus.WithFields(logrus.Fields{
			"id":             remoteResource.ResourceId(),
			"type":           remoteResource.ResourceType(),
			"network_acl_id": *remoteResource.Attrs.GetString("network_acl_id"),
		}).Debug("Ignoring default ACL rule as it is not managed by IaC")
	}

	*remoteResources = newRemoteResources

	return nil
}

func (m *AwsDefaultNetworkACLRule) isDefaultACLRule(res *resource.Resource) bool {

	isIPv4 := res.Attrs.GetString("cidr_block") != nil

	if isIPv4 {
		if number := res.Attrs.GetFloat64("rule_number"); number != nil && int(*number) != 32767 {
			return false
		}
		if cidr := res.Attrs.GetString("cidr_block"); cidr != nil && *cidr != "0.0.0.0/0" {
			return false
		}
	}

	if !isIPv4 {
		if number := res.Attrs.GetFloat64("rule_number"); number != nil && int(*number) != 32768 {
			return false
		}
		if cidr := res.Attrs.GetString("ipv6_cidr_block"); cidr != nil && *cidr != "::/0" {
			return false
		}
	}

	if action := res.Attrs.GetString("rule_action"); action != nil && *action != "deny" {
		return false
	}

	if proto := res.Attrs.GetString("protocol"); proto != nil && *proto != "-1" {
		return false
	}

	return true
}
