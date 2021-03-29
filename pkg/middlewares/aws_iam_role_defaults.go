package middlewares

import (
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

// When scanning a brand new AWS account, some users may see irrelevant results about default AWS role policies.
// We ignore these resources by default when strict mode is disabled.
type AwsIamRoleDefaults struct{}

var ignoredIamRoleIds = []string{
	"AWSServiceRoleForSSO",
	"OrganizationAccountAccessRole",
}

func NewAwsIamRoleDefaults() AwsIamRoleDefaults {
	return AwsIamRoleDefaults{}
}

func (m AwsIamRoleDefaults) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {
	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than role policy
		if remoteResource.TerraformType() != aws.AwsIamRoleResourceType {
			continue
		}

		existInState := false
		for _, stateResource := range *resourcesFromState {
			if resource.IsSameResource(remoteResource, stateResource) {
				existInState = true
				break
			}
		}

		if existInState {
			continue
		}

		for _, id := range ignoredIamRoleIds {
			if remoteResource.TerraformId() == id {
				*resourcesFromState = append(*resourcesFromState, remoteResource)

				logrus.WithFields(logrus.Fields{
					"id":   remoteResource.TerraformId(),
					"type": remoteResource.TerraformType(),
				}).Debug("Ignoring default iam role as it is not managed by IaC")
			}
		}
	}

	return nil
}
