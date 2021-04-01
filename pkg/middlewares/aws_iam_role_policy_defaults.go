package middlewares

import (
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

// When scanning a brand new AWS account, some users may see irrelevant results about default AWS role policies.
// We ignore these resources by default when strict mode is disabled.
type AwsIamRolePolicyDefaults struct{}

func NewAwsIamRolePolicyDefaults() AwsIamRolePolicyDefaults {
	return AwsIamRolePolicyDefaults{}
}

func (m AwsIamRolePolicyDefaults) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {
	newRemoteResources := make([]resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than role policy
		if remoteResource.TerraformType() != aws.AwsIamRolePolicyResourceType {
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

		if existInState {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		var role *aws.AwsIamRole
		for _, res := range *remoteResources {
			if res.TerraformType() != aws.AwsIamRoleResourceType {
				continue
			}

			if res.(*aws.AwsIamRole).Id == *remoteResource.(*aws.AwsIamRolePolicy).Role {
				role = res.(*aws.AwsIamRole)
			}
		}

		// If we couldn't find the linked role, don't ignore the resource
		if role == nil {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		match, err := filepath.Match(ignoredIamRolePathGlob, *role.Path)
		if err != nil {
			return err
		}

		if !match {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		logrus.WithFields(logrus.Fields{
			"id":   remoteResource.TerraformId(),
			"type": remoteResource.TerraformType(),
		}).Debug("Ignoring default iam role policy as it is not managed by IaC")
	}

	*remoteResources = newRemoteResources

	return nil
}
