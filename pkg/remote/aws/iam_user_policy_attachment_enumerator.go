package aws

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type IamUserPolicyAttachmentEnumerator struct {
	repository repository.IAMRepository
	factory    resource.ResourceFactory
}

func NewIamUserPolicyAttachmentEnumerator(repository repository.IAMRepository, factory resource.ResourceFactory) *IamUserPolicyAttachmentEnumerator {
	return &IamUserPolicyAttachmentEnumerator{
		repository,
		factory,
	}
}

func (e *IamUserPolicyAttachmentEnumerator) SupportedType() resource.ResourceType {
	return resourceaws.AwsIamUserPolicyAttachmentResourceType
}

func (e *IamUserPolicyAttachmentEnumerator) Enumerate() ([]*resource.Resource, error) {
	users, err := e.repository.ListAllUsers()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), resourceaws.AwsIamUserResourceType)
	}

	results := make([]*resource.Resource, 0)
	policyAttachments, err := e.repository.ListAllUserPolicyAttachments(users)
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	for _, attachedPol := range policyAttachments {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				fmt.Sprintf("%s-%s", *attachedPol.PolicyName, attachedPol.UserName),
				map[string]interface{}{
					"user":       attachedPol.UserName,
					"policy_arn": *attachedPol.PolicyArn,
				},
			),
		)
	}

	return results, nil
}
