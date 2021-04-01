package middlewares

import (
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

const ignoredIamRolePathGlob = "/aws-service-role/*"

// When scanning a brand new AWS account, some users may see irrelevant results about default AWS role policies.
// We ignore these resources by default when strict mode is disabled.
type AwsIamRoleDefaults struct{}

func NewAwsIamRoleDefaults() AwsIamRoleDefaults {
	return AwsIamRoleDefaults{}
}

func (m AwsIamRoleDefaults) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {
	newRemoteResources := make([]resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than iam role
		if remoteResource.TerraformType() != aws.AwsIamRoleResourceType {
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

		match, err := filepath.Match(ignoredIamRolePathGlob, *remoteResource.(*aws.AwsIamRole).Path)
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
		}).Debug("Ignoring default iam role as it is not managed by IaC")
	}

	*remoteResources = newRemoteResources

	return nil
}
