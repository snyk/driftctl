package middlewares

import (
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

/**
  When listing policy attachment from aws we retrieve only user_policy_attachment or role_policy_attachment thus making it
  impossible to compare with policy_attachment that could exist in terraform.
  We decided to transform all attachments to policy_attachment so we can find which attachments are managed.
*/

type IamPolicyAttachmentTransformer struct {
	resourceFactory resource.ResourceFactory
}

func NewIamPolicyAttachmentTransformer(resourceFactory resource.ResourceFactory) IamPolicyAttachmentTransformer {
	return IamPolicyAttachmentTransformer{
		resourceFactory,
	}
}

func (m IamPolicyAttachmentTransformer) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	*remoteResources = m.transform(remoteResources)
	*resourcesFromState = m.transform(resourcesFromState)
	return nil
}

func (m IamPolicyAttachmentTransformer) transform(resources *[]*resource.Resource) []*resource.Resource {
	var newResources []*resource.Resource
	for _, res := range *resources {
		if res.ResourceType() != aws.AwsIamUserPolicyAttachmentResourceType &&
			res.ResourceType() != aws.AwsIamRolePolicyAttachmentResourceType {
			newResources = append(newResources, res)
			continue
		}

		if res.ResourceType() == aws.AwsIamUserPolicyAttachmentResourceType {
			attrs := *res.Attributes()
			policyAttachmentData := resource.Attributes{
				"id":         res.ResourceId(),
				"policy_arn": attrs["policy_arn"],
				"users":      []interface{}{attrs["user"]},
				"groups":     []interface{}{},
				"roles":      []interface{}{},
			}

			policyAttachment := m.resourceFactory.CreateAbstractResource(aws.AwsIamPolicyAttachmentResourceType, res.ResourceId(), policyAttachmentData)

			newResources = append(newResources, policyAttachment)
			continue
		}

		if res.ResourceType() == aws.AwsIamRolePolicyAttachmentResourceType {
			attrs := *res.Attributes()
			policyAttachmentData := resource.Attributes{
				"id":         res.ResourceId(),
				"policy_arn": attrs["policy_arn"],
				"users":      []interface{}{},
				"groups":     []interface{}{},
				"roles":      []interface{}{attrs["role"]},
			}

			policyAttachment := m.resourceFactory.CreateAbstractResource(aws.AwsIamPolicyAttachmentResourceType, res.ResourceId(), policyAttachmentData)

			newResources = append(newResources, policyAttachment)
			continue
		}
	}
	return newResources
}
