package middlewares

import (
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// AwsIAMPolicyAttachmentTransformer is a simple middleware to recreate remote attachments to match ones from
// the state if they do exist.
type AwsIAMPolicyAttachmentTransformer struct {
	resourceFactory resource.ResourceFactory
}

func NewAwsIAMPolicyAttachmentTransformer(resourceFactory resource.ResourceFactory) AwsIAMPolicyAttachmentTransformer {
	return AwsIAMPolicyAttachmentTransformer{
		resourceFactory: resourceFactory,
	}
}

func (m AwsIAMPolicyAttachmentTransformer) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	newRemoteResources := make([]*resource.Resource, 0, len(*remoteResources))

	for _, remoteRes := range *remoteResources {
		if remoteRes.ResourceType() != aws.AwsIamGroupPolicyAttachmentResourceType {
			newRemoteResources = append(newRemoteResources, remoteRes)
			continue
		}

		var newId string

		for _, stateRes := range *resourcesFromState {
			if stateRes.ResourceType() != aws.AwsIamGroupPolicyAttachmentResourceType {
				continue
			}

			statePolicyArn := stateRes.Attributes().GetString("policy_arn")
			remotePolicyArn := remoteRes.Attributes().GetString("policy_arn")
			if statePolicyArn == nil || remotePolicyArn == nil {
				break
			}

			if *statePolicyArn == *remotePolicyArn {
				newId = stateRes.ResourceId()
				break
			}
		}

		// If we didn't manage to find the resource in state, let it appear as unmanaged
		if newId == "" {
			newRemoteResources = append(newRemoteResources, remoteRes)
			continue
		}

		newRemoteResources = append(newRemoteResources, m.resourceFactory.CreateAbstractResource(
			aws.AwsIamGroupPolicyAttachmentResourceType,
			newId,
			*remoteRes.Attributes(),
		))
	}

	*remoteResources = newRemoteResources
	return nil
}
