package aws

import (
	"fmt"

	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	resourceaws "github.com/snyk/driftctl/pkg/resource/aws"
)

type IamGroupPolicyAttachmentEnumerator struct {
	repository repository.IAMRepository
	factory    resource.ResourceFactory
}

func NewIamGroupPolicyAttachmentEnumerator(repository repository.IAMRepository, factory resource.ResourceFactory) *IamGroupPolicyAttachmentEnumerator {
	return &IamGroupPolicyAttachmentEnumerator{
		repository,
		factory,
	}
}

func (e *IamGroupPolicyAttachmentEnumerator) SupportedType() resource.ResourceType {
	return resourceaws.AwsIamGroupPolicyAttachmentResourceType
}

func (e *IamGroupPolicyAttachmentEnumerator) Enumerate() ([]*resource.Resource, error) {
	groups, err := e.repository.ListAllGroups()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), resourceaws.AwsIamGroupResourceType)
	}

	results := make([]*resource.Resource, 0)

	policyAttachments, err := e.repository.ListAllGroupPolicyAttachments(groups)
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	for _, attachedPol := range policyAttachments {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				fmt.Sprintf("%s-%s", *attachedPol.PolicyName, attachedPol.GroupName),
				map[string]interface{}{
					"group":      attachedPol.GroupName,
					"policy_arn": *attachedPol.PolicyArn,
				},
			),
		)
	}

	return results, nil
}
