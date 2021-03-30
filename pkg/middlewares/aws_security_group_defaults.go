package middlewares

import (
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

// When scanning a brand new AWS account, some users may see irrelevant results about default AWS role policies.
// We ignore these resources by default when strict mode is disabled.
type AwsSecurityGroupRuleDefaults struct{}

func NewAwsSecurityGroupRuleDefaults() AwsSecurityGroupRuleDefaults {
	return AwsSecurityGroupRuleDefaults{}
}

func (m AwsSecurityGroupRuleDefaults) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {
	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than iam role
		if remoteResource.TerraformType() != aws.AwsSecurityGroupResourceType {
			continue
		}

		existInState := false
		for _, stateResource := range *resourcesFromState {
			if resource.IsSameResource(remoteResource, stateResource) {
				existInState = true
				break
			}
		}

		if existInState || *remoteResource.(*aws.AwsSecurityGroup).Name != "default" {
			continue
		}

		*resourcesFromState = append(*resourcesFromState, remoteResource)

		logrus.WithFields(logrus.Fields{
			"id":   remoteResource.TerraformId(),
			"type": remoteResource.TerraformType(),
		}).Debug("Ignoring default aws security group as it is not managed by IaC")
	}

	return nil
}
