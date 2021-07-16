package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
)

var iamRoleExclusionList = map[string]struct{}{
	// Enabled by default for aws to enable support, not removable
	"AWSServiceRoleForSupport": {},
	// Enabled and not removable for every org account
	"AWSServiceRoleForOrganizations": {},
	// Not manageable by IaC and set by default
	"AWSServiceRoleForTrustedAdvisor": {},
}

type IamRoleEnumerator struct {
	repository repository.IAMRepository
	factory    resource.ResourceFactory
}

func NewIamRoleEnumerator(repository repository.IAMRepository, factory resource.ResourceFactory) *IamRoleEnumerator {
	return &IamRoleEnumerator{
		repository,
		factory,
	}
}

func (e *IamRoleEnumerator) SupportedType() resource.ResourceType {
	return resourceaws.AwsIamRoleResourceType
}

func awsIamRoleShouldBeIgnored(roleName string) bool {
	_, ok := iamRoleExclusionList[roleName]
	return ok
}

func (e *IamRoleEnumerator) Enumerate() ([]resource.Resource, error) {
	roles, err := e.repository.ListAllRoles()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, string(e.SupportedType()))
	}

	results := make([]resource.Resource, 0)
	for _, role := range roles {
		if role.RoleName != nil && awsIamRoleShouldBeIgnored(*role.RoleName) {
			continue
		}

		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*role.RoleName,
				map[string]interface{}{
					"path": *role.Path,
				},
			),
		)
	}

	return results, nil
}
