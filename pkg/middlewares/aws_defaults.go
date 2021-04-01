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
type AwsDefaults struct{}

func NewAwsDefaults() AwsDefaults {
	return AwsDefaults{}
}

func awsIamRoleDefaults(remoteResources, resourcesFromState *[]resource.Resource) error {
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

func awsIamPolicyAttachmentDefaults(remoteResources, resourcesFromState *[]resource.Resource) error {
	newRemoteResources := make([]resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than iam policy attachment
		if remoteResource.TerraformType() != aws.AwsIamPolicyAttachmentResourceType {
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

		for _, roleId := range *remoteResource.(*aws.AwsIamPolicyAttachment).Roles {
			var role *aws.AwsIamRole
			for _, res := range *remoteResources {
				if res.TerraformType() != aws.AwsIamRoleResourceType {
					continue
				}

				if res.(*aws.AwsIamRole).Id == roleId {
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
		}

		logrus.WithFields(logrus.Fields{
			"id":   remoteResource.TerraformId(),
			"type": remoteResource.TerraformType(),
		}).Debug("Ignoring default iam policy attachment as it is not managed by IaC")
	}

	*remoteResources = newRemoteResources

	return nil
}

func awsIamRolePolicyDefaults(remoteResources, resourcesFromState *[]resource.Resource) error {
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

func (m AwsDefaults) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {
	err := awsIamRoleDefaults(remoteResources, resourcesFromState)
	if err != nil {
		return err
	}

	err = awsIamPolicyAttachmentDefaults(remoteResources, resourcesFromState)
	if err != nil {
		return err
	}

	err = awsIamRolePolicyDefaults(remoteResources, resourcesFromState)
	if err != nil {
		return err
	}

	return nil
}
