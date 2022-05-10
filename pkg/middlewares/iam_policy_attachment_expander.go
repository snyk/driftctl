package middlewares

import (
	"fmt"

	"github.com/snyk/driftctl/pkg/resource"
	resourceaws "github.com/snyk/driftctl/pkg/resource/aws"
)

// Split Policy attachment when there is multiple user and groups and generate a repeatable id
type IamPolicyAttachmentExpander struct {
	resourceFactory resource.ResourceFactory
}

func NewIamPolicyAttachmentExpander(resourceFactory resource.ResourceFactory) IamPolicyAttachmentExpander {
	return IamPolicyAttachmentExpander{
		resourceFactory,
	}
}

func (m IamPolicyAttachmentExpander) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	var newStateResources = make([]*resource.Resource, 0)

	for _, stateResource := range *resourcesFromState {
		// Ignore all resources other than policy attachment
		if stateResource.ResourceType() != resourceaws.AwsIamPolicyAttachmentResourceType {
			newStateResources = append(newStateResources, stateResource)
			continue
		}

		newStateResources = append(newStateResources, m.expand(stateResource)...)
	}

	var newRemoteResources = make([]*resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than policy attachment
		if remoteResource.ResourceType() != resourceaws.AwsIamPolicyAttachmentResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		newRemoteResources = append(newRemoteResources, m.expand(remoteResource)...)
	}

	*resourcesFromState = newStateResources
	*remoteResources = newRemoteResources

	return nil
}

func (m IamPolicyAttachmentExpander) expand(policyAttachment *resource.Resource) []*resource.Resource {
	var newResources []*resource.Resource
	users := policyAttachment.Attrs.GetSlice("users")
	// we create one attachment per user
	for _, user := range users {
		user := user.(string)
		newAttachment := m.resourceFactory.CreateAbstractResource(
			resourceaws.AwsIamPolicyAttachmentResourceType,
			fmt.Sprintf("%s-%s", user, (*policyAttachment.Attrs)["policy_arn"]),
			map[string]interface{}{
				"policy_arn": *policyAttachment.Attrs.GetString("policy_arn"),
				"users":      []interface{}{user},
			},
		)
		newResources = append(newResources, newAttachment)
	}

	roles := policyAttachment.Attrs.GetSlice("roles")
	// we create one attachment per role
	for _, role := range roles {
		role := role.(string)
		newAttachment := m.resourceFactory.CreateAbstractResource(
			resourceaws.AwsIamPolicyAttachmentResourceType,
			fmt.Sprintf("%s-%s", role, (*policyAttachment.Attrs)["policy_arn"]),
			map[string]interface{}{
				"policy_arn": *policyAttachment.Attrs.GetString("policy_arn"),
				"roles":      []interface{}{role},
			},
		)
		newResources = append(newResources, newAttachment)
	}

	groups := policyAttachment.Attrs.GetSlice("groups")
	// we create one attachment per group
	for _, group := range groups {
		group := group.(string)
		newAttachment := m.resourceFactory.CreateAbstractResource(
			resourceaws.AwsIamPolicyAttachmentResourceType,
			fmt.Sprintf("%s-%s", group, (*policyAttachment.Attrs)["policy_arn"]),
			map[string]interface{}{
				"policy_arn": *policyAttachment.Attrs.GetString("policy_arn"),
				"groups":     []interface{}{group},
			},
		)
		newResources = append(newResources, newAttachment)
	}
	return newResources
}
