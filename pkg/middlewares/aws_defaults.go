package middlewares

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

const defaultIamRolePathPrefix = "/aws-service-role/"

// AwsDefaults represents service-linked AWS resources
// When scanning a AWS account, some users may see irrelevant results about default AWS roles or role policies.
// We ignore these resources by default when strict mode is disabled.
type AwsDefaults struct{}

func NewAwsDefaults() AwsDefaults {
	return AwsDefaults{}
}

func (m AwsDefaults) awsIamRoleDefaults(remoteResources []*resource.Resource) []*resource.Resource {
	resourcesToIgnore := make([]*resource.Resource, 0)

	for _, remoteResource := range remoteResources {
		// Ignore all resources other than iam role
		if remoteResource.ResourceType() != aws.AwsIamRoleResourceType {
			continue
		}

		path := remoteResource.Attributes().GetString("path")
		if path == nil {
			continue
		}

		if match := strings.HasPrefix(*path, defaultIamRolePathPrefix); match {
			resourcesToIgnore = append(resourcesToIgnore, remoteResource)
		}
	}

	return resourcesToIgnore
}

func (m AwsDefaults) awsIamPolicyAttachmentDefaults(remoteResources []*resource.Resource) []*resource.Resource {
	resourcesToIgnore := make([]*resource.Resource, 0)

	for _, remoteResource := range remoteResources {
		// Ignore all resources other than iam policy attachment
		if remoteResource.ResourceType() != aws.AwsIamPolicyAttachmentResourceType {
			continue
		}

		// NewIamPolicyAttachmentExpander ensures that each attachment resource has only one user, group, or role
		if (remoteResource.Attrs.GetSlice("users") != nil) || (remoteResource.Attrs.GetSlice("groups") != nil) {
			continue
		}

		roleId := remoteResource.Attrs.GetSlice("roles")[0]
		for _, res := range remoteResources {
			if res.ResourceType() == aws.AwsIamRoleResourceType && res.Id == roleId {
				rolePath := res.Attributes().GetString("path")
				if match := strings.HasPrefix(*rolePath, defaultIamRolePathPrefix); match {
					resourcesToIgnore = append(resourcesToIgnore, remoteResource)
				}
				break
			}
		}
	}

	return resourcesToIgnore
}

func (m AwsDefaults) awsIamRolePolicyDefaults(remoteResources []*resource.Resource) []*resource.Resource {
	resourcesToIgnore := make([]*resource.Resource, 0)

	for _, remoteResource := range remoteResources {
		// Ignore all resources other than role policy
		if remoteResource.ResourceType() != aws.AwsIamRolePolicyResourceType {
			continue
		}

		var role *resource.Resource
		for _, res := range remoteResources {
			if res.ResourceType() == aws.AwsIamRoleResourceType &&
				res.ResourceId() == (*remoteResource.Attrs)["role"] {
				role = res
				break
			}
		}

		if role == nil {
			logrus.Warnf("Role for %s role policy not found. Is that supposed to happen ?", remoteResource.ResourceId())
			continue
		}

		if match := strings.HasPrefix((*role.Attrs)["path"].(string), defaultIamRolePathPrefix); match {
			resourcesToIgnore = append(resourcesToIgnore, remoteResource)
		}
	}

	return resourcesToIgnore
}

func (m AwsDefaults) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	newRemoteResources := make([]*resource.Resource, 0)
	newResourcesFromState := make([]*resource.Resource, 0)
	resourcesToIgnore := make([]*resource.Resource, 0)

	resourcesToIgnore = append(resourcesToIgnore, m.awsIamRoleDefaults(*remoteResources)...)
	resourcesToIgnore = append(resourcesToIgnore, m.awsIamPolicyAttachmentDefaults(*remoteResources)...)
	resourcesToIgnore = append(resourcesToIgnore, m.awsIamRolePolicyDefaults(*remoteResources)...)

	for _, res := range *remoteResources {
		ignored := false

		for _, resourceToIgnore := range resourcesToIgnore {
			if res.Equal(resourceToIgnore) {
				ignored = true
				break
			}
		}

		if !ignored {
			newRemoteResources = append(newRemoteResources, res)
			continue
		}

		logrus.WithFields(logrus.Fields{
			"id":   res.ResourceId(),
			"type": res.ResourceType(),
		}).Debug("Ignoring default AWS resource")
	}

	for _, res := range *resourcesFromState {
		ignored := false

		for _, resourceToIgnore := range resourcesToIgnore {
			if res.Equal(resourceToIgnore) {
				ignored = true
				break
			}
		}

		if !ignored {
			newResourcesFromState = append(newResourcesFromState, res)
			continue
		}

		logrus.WithFields(logrus.Fields{
			"id":   res.ResourceId(),
			"type": res.ResourceType(),
		}).Debug("Ignoring default AWS resource")
	}

	*remoteResources = newRemoteResources
	*resourcesFromState = newResourcesFromState

	return nil
}
