package aws

import (
	"fmt"

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

func (e *IamRolePolicyEnumerator) Enumerate() ([]*resource.Resource, error) {
	roles, err := e.repository.ListAllRoles()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), resourceaws.AwsIamRoleResourceType)
	}

	policies, err := e.repository.ListAllRolePolicies(roles)
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(policies))
	for _, policy := range policies {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				fmt.Sprintf("%s:%s", policy.RoleName, policy.Policy),
				map[string]interface{}{
					"role": policy.RoleName,
				},
			),
		)
	}

	return results, nil
}
