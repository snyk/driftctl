package repository

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
)

type IAMRepository interface {
	ListAllAccessKeys([]*iam.User) ([]*iam.AccessKeyMetadata, error)
	ListAllUsers() ([]*iam.User, error)
	ListAllPolicies() ([]*iam.Policy, error)
	ListAllRoles() ([]*iam.Role, error)
	ListAllRolePolicyAttachments([]*iam.Role) ([]*AttachedRolePolicy, error)
	ListAllRolePolicies([]*iam.Role) ([]string, error)
	ListAllUserPolicyAttachments([]*iam.User) ([]*AttachedUserPolicy, error)
	ListAllUserPolicies([]*iam.User) ([]string, error)
}

type iamRepository struct {
	client iamiface.IAMAPI
}

func NewIAMRepository(session *session.Session) *iamRepository {
	return &iamRepository{
		iam.New(session),
	}
}

func (r *iamRepository) ListAllAccessKeys(users []*iam.User) ([]*iam.AccessKeyMetadata, error) {
	var resources []*iam.AccessKeyMetadata
	for _, user := range users {
		input := &iam.ListAccessKeysInput{
			UserName: user.UserName,
		}
		err := r.client.ListAccessKeysPages(input, func(res *iam.ListAccessKeysOutput, lastPage bool) bool {
			resources = append(resources, res.AccessKeyMetadata...)
			return !lastPage
		})
		if err != nil {
			return nil, err
		}
	}

	return resources, nil
}

func (r *iamRepository) ListAllUsers() ([]*iam.User, error) {
	var resources []*iam.User
	input := &iam.ListUsersInput{}
	err := r.client.ListUsersPages(input, func(res *iam.ListUsersOutput, lastPage bool) bool {
		resources = append(resources, res.Users...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return resources, nil
}

func (r *iamRepository) ListAllPolicies() ([]*iam.Policy, error) {
	var resources []*iam.Policy
	input := &iam.ListPoliciesInput{
		Scope: aws.String(iam.PolicyScopeTypeLocal),
	}
	err := r.client.ListPoliciesPages(input, func(res *iam.ListPoliciesOutput, lastPage bool) bool {
		resources = append(resources, res.Policies...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return resources, nil
}

func (r *iamRepository) ListAllRoles() ([]*iam.Role, error) {
	var resources []*iam.Role
	input := &iam.ListRolesInput{}
	err := r.client.ListRolesPages(input, func(res *iam.ListRolesOutput, lastPage bool) bool {
		resources = append(resources, res.Roles...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return resources, nil
}

func (r *iamRepository) ListAllRolePolicyAttachments(roles []*iam.Role) ([]*AttachedRolePolicy, error) {
	var resources []*AttachedRolePolicy
	for _, role := range roles {
		input := &iam.ListAttachedRolePoliciesInput{
			RoleName: role.RoleName,
		}
		err := r.client.ListAttachedRolePoliciesPages(input, func(res *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool {
			for _, policy := range res.AttachedPolicies {
				p := *policy
				resources = append(resources, &AttachedRolePolicy{
					AttachedPolicy: p,
					RoleName:       *input.RoleName,
				})
			}
			return !lastPage
		})
		if err != nil {
			return nil, err
		}
	}

	return resources, nil
}

func (r *iamRepository) ListAllRolePolicies(roles []*iam.Role) ([]string, error) {
	var resources []string
	for _, role := range roles {
		input := &iam.ListRolePoliciesInput{
			RoleName: role.RoleName,
		}
		err := r.client.ListRolePoliciesPages(input, func(res *iam.ListRolePoliciesOutput, lastPage bool) bool {
			for _, policy := range res.PolicyNames {
				resources = append(resources, fmt.Sprintf("%s:%s", *input.RoleName, *policy))
			}
			return !lastPage
		})
		if err != nil {
			return nil, err
		}
	}

	return resources, nil
}

func (r *iamRepository) ListAllUserPolicyAttachments(users []*iam.User) ([]*AttachedUserPolicy, error) {
	var resources []*AttachedUserPolicy
	for _, user := range users {
		input := &iam.ListAttachedUserPoliciesInput{
			UserName: user.UserName,
		}
		err := r.client.ListAttachedUserPoliciesPages(input, func(res *iam.ListAttachedUserPoliciesOutput, lastPage bool) bool {
			for _, policy := range res.AttachedPolicies {
				p := *policy
				resources = append(resources, &AttachedUserPolicy{
					AttachedPolicy: p,
					UserName:       *input.UserName,
				})
			}
			return !lastPage
		})
		if err != nil {
			return nil, err
		}
	}

	return resources, nil
}

func (r *iamRepository) ListAllUserPolicies(users []*iam.User) ([]string, error) {
	var resources []string
	for _, user := range users {
		input := &iam.ListUserPoliciesInput{
			UserName: user.UserName,
		}
		err := r.client.ListUserPoliciesPages(input, func(res *iam.ListUserPoliciesOutput, lastPage bool) bool {
			for _, polName := range res.PolicyNames {
				resources = append(resources, fmt.Sprintf("%s:%s", *input.UserName, *polName))
			}
			return !lastPage
		})
		if err != nil {
			return nil, err
		}
	}

	return resources, nil
}

type AttachedUserPolicy struct {
	iam.AttachedPolicy
	UserName string
}

type AttachedRolePolicy struct {
	iam.AttachedPolicy
	RoleName string
}
