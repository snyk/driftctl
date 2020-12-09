package middlewares

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
)

// Split Policy attachment when there is multiple user and groups and generate a repeatable id
type IamPolicyAttachmentSanitizer struct{}

func NewIamPolicyAttachmentSanitizer() IamPolicyAttachmentSanitizer {
	return IamPolicyAttachmentSanitizer{}
}

func (m IamPolicyAttachmentSanitizer) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {
	var newStateResources = make([]resource.Resource, 0)

	for _, stateResource := range *resourcesFromState {
		// Ignore all resources other than policy attachment
		if stateResource.TerraformType() != resourceaws.AwsIamPolicyAttachmentResourceType {
			newStateResources = append(newStateResources, stateResource)
			continue
		}

		policyAttachment := stateResource.(*resourceaws.AwsIamPolicyAttachment)

		newStateResources = append(newStateResources, m.sanitize(policyAttachment)...)
	}

	var newRemoteResources = make([]resource.Resource, 0)

	for _, stateResource := range *remoteResources {
		// Ignore all resources other than policy attachment
		if stateResource.TerraformType() != resourceaws.AwsIamPolicyAttachmentResourceType {
			newRemoteResources = append(newRemoteResources, stateResource)
			continue
		}

		policyAttachment := stateResource.(*resourceaws.AwsIamPolicyAttachment)

		newRemoteResources = append(newRemoteResources, m.sanitize(policyAttachment)...)
	}

	*resourcesFromState = newStateResources
	*remoteResources = newRemoteResources

	return nil
}

func (m IamPolicyAttachmentSanitizer) sanitize(policyAttachment *resourceaws.AwsIamPolicyAttachment) []resource.Resource {
	newResources := make([]resource.Resource, 0, len(policyAttachment.Users))

	// we create one attachment per user
	for _, user := range policyAttachment.Users {
		newAttachment := *policyAttachment

		// Id is generated with unique id in state so we override it with something repeatable
		newAttachment.Id = fmt.Sprintf("%s-%s", user, *policyAttachment.PolicyArn)

		newAttachment.Users = []string{user}
		newResources = append(newResources, &newAttachment)
	}

	// we create one attachment per role
	for _, role := range policyAttachment.Roles {
		newAttachment := *policyAttachment

		// Id is generated with unique id in state so we override it with something repeatable
		newAttachment.Id = fmt.Sprintf("%s-%s", role, *policyAttachment.PolicyArn)

		newAttachment.Roles = []string{role}
		newResources = append(newResources, &newAttachment)
	}
	return newResources
}
