package middlewares

import (
	"strings"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

const defaultIamRolePathPrefix = "/aws-service-role/"

// AwsDefaults represents service-linked AWS resources
// When scanning a AWS account, some users may see irrelevant results about default AWS roles or role policies.
// We ignore these resources by default when strict mode is disabled.
type AwsDefaults struct{}

func NewAwsDefaults() AwsDefaults {
	return AwsDefaults{}
}

func (m AwsDefaults) awsIamRoleDefaults(remoteResources []resource.Resource) []resource.Resource {
	resourcesToIgnore := make([]resource.Resource, 0)

	for _, remoteResource := range remoteResources {
		// Ignore all resources other than iam role
		if remoteResource.TerraformType() != aws.AwsIamRoleResourceType {
			continue
		}

		if match := strings.HasPrefix(*remoteResource.(*aws.AwsIamRole).Path, defaultIamRolePathPrefix); match {
			resourcesToIgnore = append(resourcesToIgnore, remoteResource)
		}
	}

	return resourcesToIgnore
}

func (m AwsDefaults) awsIamPolicyAttachmentDefaults(remoteResources []resource.Resource) []resource.Resource {
	resourcesToIgnore := make([]resource.Resource, 0)

	for _, remoteResource := range remoteResources {
		// Ignore all resources other than iam policy attachment
		if remoteResource.TerraformType() != aws.AwsIamPolicyAttachmentResourceType {
			continue
		}

		defaultRolesCount := 0
		for _, roleId := range *remoteResource.(*aws.AwsIamPolicyAttachment).Roles {
			var role *aws.AwsIamRole
			for _, res := range remoteResources {
				if res.TerraformType() == aws.AwsIamRoleResourceType && res.TerraformId() == roleId {
					role = res.(*aws.AwsIamRole)
					break
				}
			}

			if match := strings.HasPrefix(*role.Path, defaultIamRolePathPrefix); match {
				defaultRolesCount++
			}
		}

		// Check if all of the policy's roles are default AWS roles
		if defaultRolesCount == len(*remoteResource.(*aws.AwsIamPolicyAttachment).Roles) {
			resourcesToIgnore = append(resourcesToIgnore, remoteResource)
		}
	}

	return resourcesToIgnore
}

func (m AwsDefaults) awsIamRolePolicyDefaults(remoteResources []resource.Resource) []resource.Resource {
	resourcesToIgnore := make([]resource.Resource, 0)

	for _, remoteResource := range remoteResources {
		// Ignore all resources other than role policy
		if remoteResource.TerraformType() != aws.AwsIamRolePolicyResourceType {
			continue
		}

		var role *aws.AwsIamRole
		for _, res := range remoteResources {
			if res.TerraformType() == aws.AwsIamRoleResourceType && res.TerraformId() == *remoteResource.(*aws.AwsIamRolePolicy).Role {
				role = res.(*aws.AwsIamRole)
				break
			}
		}

		if match := strings.HasPrefix(*role.Path, defaultIamRolePathPrefix); match {
			resourcesToIgnore = append(resourcesToIgnore, remoteResource)
		}
	}

	return resourcesToIgnore
}

func (m AwsDefaults) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {
	newRemoteResources := make([]resource.Resource, 0)
	newResourcesFromState := make([]resource.Resource, 0)
	resourcesToIgnore := make([]resource.Resource, 0)

	resourcesToIgnore = append(resourcesToIgnore, m.awsIamRoleDefaults(*remoteResources)...)
	resourcesToIgnore = append(resourcesToIgnore, m.awsIamPolicyAttachmentDefaults(*remoteResources)...)
	resourcesToIgnore = append(resourcesToIgnore, m.awsIamRolePolicyDefaults(*remoteResources)...)

	for _, res := range *remoteResources {
		ignored := false

		for _, resourceToIgnore := range resourcesToIgnore {
			if resource.IsSameResource(res, resourceToIgnore) {
				ignored = true
				break
			}
		}

		if !ignored {
			newRemoteResources = append(newRemoteResources, res)
			continue
		}

		logrus.WithFields(logrus.Fields{
			"id":   res.TerraformId(),
			"type": res.TerraformType(),
		}).Debug("Ignoring default AWS resource")
	}

	for _, res := range *resourcesFromState {
		ignored := false

		for _, resourceToIgnore := range resourcesToIgnore {
			if resource.IsSameResource(res, resourceToIgnore) {
				ignored = true
				break
			}
		}

		if !ignored {
			newResourcesFromState = append(newResourcesFromState, res)
			continue
		}

		logrus.WithFields(logrus.Fields{
			"id":   res.TerraformId(),
			"type": res.TerraformType(),
		}).Debug("Ignoring default AWS resource")
	}

	*remoteResources = newRemoteResources
	*resourcesFromState = newResourcesFromState

	return nil
}
