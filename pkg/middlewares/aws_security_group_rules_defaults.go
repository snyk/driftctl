package middlewares

import (
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

// When scanning a brand new AWS account, some users may see irrelevant results about default AWS security group rules.
// We ignore these resources by default when strict mode is disabled.
type AwsSecurityGroupRuleDefaults struct{}

func NewAwsSecurityGroupRuleDefaults() AwsSecurityGroupRuleDefaults {
	return AwsSecurityGroupRuleDefaults{}
}

func (m AwsSecurityGroupRuleDefaults) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {
	newRemoteResources := make([]resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than security group rule
		if remoteResource.TerraformType() != aws.AwsSecurityGroupRuleResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		existInState := false
		for _, stateResource := range *resourcesFromState {
			if resource.IsSameResource(remoteResource, stateResource) {
				existInState = true
				break
			}
		}

		if existInState || !isDefaultSecurityGroupRule(remoteResource.(*aws.AwsSecurityGroupRule), *remoteResources) {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		logrus.WithFields(logrus.Fields{
			"id":   remoteResource.TerraformId(),
			"type": remoteResource.TerraformType(),
		}).Debug("Ignoring default aws security group rule as it is not managed by IaC")
	}

	*remoteResources = newRemoteResources

	return nil
}

// Return true if the record is considered as default one added by aws
func isDefaultSecurityGroupRule(sgr *aws.AwsSecurityGroupRule, remoteResources []resource.Resource) bool {
	isDefaultSecurityGroup := false
	for _, res := range remoteResources {
		if res.TerraformType() != aws.AwsSecurityGroupResourceType {
			continue
		}

		if res.TerraformId() != *sgr.SecurityGroupId {
			continue
		}

		if *res.(*aws.AwsSecurityGroup).Name == defaultAwsSecurityGroupName {
			isDefaultSecurityGroup = true
		}
	}

	return isDefaultSecurityGroup && *sgr.Protocol == "All" && *sgr.Type == "ingress"
}
