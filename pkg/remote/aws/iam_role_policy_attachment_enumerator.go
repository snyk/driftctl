package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type IamRolePolicyAttachmentEnumerator struct {
	repository repository.IAMRepository
	factory    resource.ResourceFactory
}

func NewIamRolePolicyAttachmentEnumerator(repository repository.IAMRepository, factory resource.ResourceFactory) *IamRolePolicyAttachmentEnumerator {
	return &IamRolePolicyAttachmentEnumerator{
		repository,
		factory,
	}
}

func (e *IamRolePolicyAttachmentEnumerator) SupportedType() resource.ResourceType {
	return resourceaws.AwsIamRolePolicyAttachmentResourceType
}

func (e *IamRolePolicyAttachmentEnumerator) Enumerate() ([]*resource.Resource, error) {
	roles, err := e.repository.ListAllRoles()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), resourceaws.AwsIamRoleResourceType)
	}

	results := make([]*resource.Resource, 0)
	rolesNotIgnored := make([]*iam.Role, 0)

	for _, role := range roles {
		if role.RoleName != nil && awsIamRoleShouldBeIgnored(*role.RoleName) {
			continue
		}
		rolesNotIgnored = append(rolesNotIgnored, role)
	}

	if len(rolesNotIgnored) == 0 {
		return results, nil
	}

	policyAttachments, err := e.repository.ListAllRolePolicyAttachments(rolesNotIgnored)
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	for _, attachedPol := range policyAttachments {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				fmt.Sprintf("%s-%s", *attachedPol.PolicyName, attachedPol.RoleName),
				map[string]interface{}{
					"role":       attachedPol.RoleName,
					"policy_arn": *attachedPol.PolicyArn,
				},
			),
		)
	}

	return results, nil
}
