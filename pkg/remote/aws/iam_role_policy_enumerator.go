package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type IamRolePolicyEnumerator struct {
	repository repository.IAMRepository
	factory    resource.ResourceFactory
}

func NewIamRolePolicyEnumerator(repository repository.IAMRepository, factory resource.ResourceFactory) *IamRolePolicyEnumerator {
	return &IamRolePolicyEnumerator{
		repository,
		factory,
	}
}

func (e *IamRolePolicyEnumerator) SupportedType() resource.ResourceType {
	return resourceaws.AwsIamRolePolicyResourceType
}

func (e *IamRolePolicyEnumerator) Enumerate() ([]resource.Resource, error) {
	roles, err := e.repository.ListAllRoles()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, resourceaws.AwsIamRolePolicyResourceType, resourceaws.AwsIamRoleResourceType)
	}

	policies, err := e.repository.ListAllRolePolicies(roles)
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsIamRolePolicyResourceType)
	}

	results := make([]resource.Resource, len(policies))
	for _, policy := range policies {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				policy,
				map[string]interface{}{},
			),
		)
	}

	return results, nil
}
