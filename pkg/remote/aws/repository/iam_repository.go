package repository

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/snyk/driftctl/pkg/remote/cache"
)

type IAMRepository interface {
	ListAllAccessKeys([]*iam.User) ([]*iam.AccessKeyMetadata, error)
	ListAllUsers() ([]*iam.User, error)
	ListAllPolicies() ([]*iam.Policy, error)
	ListAllRoles() ([]*iam.Role, error)
	ListAllRolePolicyAttachments([]*iam.Role) ([]*AttachedRolePolicy, error)
	ListAllRolePolicies([]*iam.Role) ([]RolePolicy, error)
	ListAllUserPolicyAttachments([]*iam.User) ([]*AttachedUserPolicy, error)
	ListAllUserPolicies([]*iam.User) ([]string, error)
}

type iamRepository struct {
	client iamiface.IAMAPI
	cache  cache.Cache
}

func NewIAMRepository(session *session.Session, c cache.Cache) *iamRepository {
	return &iamRepository{
		iam.New(session),
		c,
	}
}

func (r *iamRepository) ListAllAccessKeys(users []*iam.User) ([]*iam.AccessKeyMetadata, error) {
	var resources []*iam.AccessKeyMetadata
	for _, user := range users {
		cacheKey := fmt.Sprintf("iamListAllAccessKeys_user_%s", *user.UserName)
		if v := r.cache.Get(cacheKey); v != nil {
			resources = append(resources, v.([]*iam.AccessKeyMetadata)...)
			continue
		}

		userResources := make([]*iam.AccessKeyMetadata, 0)
		input := &iam.ListAccessKeysInput{
			UserName: user.UserName,
		}
		err := r.client.ListAccessKeysPages(input, func(res *iam.ListAccessKeysOutput, lastPage bool) bool {
			userResources = append(userResources, res.AccessKeyMetadata...)
			return !lastPage
		})
		if err != nil {
			return nil, err
		}

		r.cache.Put(cacheKey, userResources)
		resources = append(resources, userResources...)
	}

	return resources, nil
}

func (r *iamRepository) ListAllUsers() ([]*iam.User, error) {

	cacheKey := "iamListAllUsers"
	v := r.cache.GetAndLock(cacheKey)
	r.cache.Unlock(cacheKey)
	if v != nil {
		return v.([]*iam.User), nil
	}

	var resources []*iam.User
	input := &iam.ListUsersInput{}
	err := r.client.ListUsersPages(input, func(res *iam.ListUsersOutput, lastPage bool) bool {
		resources = append(resources, res.Users...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, resources)
	return resources, nil
}

func (r *iamRepository) ListAllPolicies() ([]*iam.Policy, error) {
	if v := r.cache.Get("iamListAllPolicies"); v != nil {
		return v.([]*iam.Policy), nil
	}

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

	r.cache.Put("iamListAllPolicies", resources)
	return resources, nil
}

func (r *iamRepository) ListAllRoles() ([]*iam.Role, error) {
	cacheKey := "iamListAllRoles"
	v := r.cache.GetAndLock(cacheKey)
	r.cache.Unlock(cacheKey)
	if v != nil {
		return v.([]*iam.Role), nil
	}

	var resources []*iam.Role
	input := &iam.ListRolesInput{}
	err := r.client.ListRolesPages(input, func(res *iam.ListRolesOutput, lastPage bool) bool {
		resources = append(resources, res.Roles...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, resources)
	return resources, nil
}

func (r *iamRepository) ListAllRolePolicyAttachments(roles []*iam.Role) ([]*AttachedRolePolicy, error) {
	var resources []*AttachedRolePolicy
	for _, role := range roles {
		cacheKey := fmt.Sprintf("iamListAllRolePolicyAttachments_role_%s", *role.RoleName)
		if v := r.cache.Get(cacheKey); v != nil {
			resources = append(resources, v.([]*AttachedRolePolicy)...)
			continue
		}

		roleResources := make([]*AttachedRolePolicy, 0)
		input := &iam.ListAttachedRolePoliciesInput{
			RoleName: role.RoleName,
		}
		err := r.client.ListAttachedRolePoliciesPages(input, func(res *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool {
			for _, policy := range res.AttachedPolicies {
				p := *policy
				roleResources = append(roleResources, &AttachedRolePolicy{
					AttachedPolicy: p,
					RoleName:       *input.RoleName,
				})
			}
			return !lastPage
		})
		if err != nil {
			return nil, err
		}

		r.cache.Put(cacheKey, roleResources)
		resources = append(resources, roleResources...)
	}

	return resources, nil
}

func (r *iamRepository) ListAllRolePolicies(roles []*iam.Role) ([]RolePolicy, error) {
	var resources []RolePolicy
	for _, role := range roles {
		cacheKey := fmt.Sprintf("iamListAllRolePolicies_role_%s", *role.RoleName)
		if v := r.cache.Get(cacheKey); v != nil {
			resources = append(resources, v.([]RolePolicy)...)
			continue
		}

		roleResources := make([]RolePolicy, 0)
		input := &iam.ListRolePoliciesInput{
			RoleName: role.RoleName,
		}
		err := r.client.ListRolePoliciesPages(input, func(res *iam.ListRolePoliciesOutput, lastPage bool) bool {
			for _, policy := range res.PolicyNames {
				roleResources = append(roleResources, RolePolicy{*policy, *input.RoleName})
			}
			return !lastPage
		})
		if err != nil {
			return nil, err
		}

		r.cache.Put(cacheKey, roleResources)
		resources = append(resources, roleResources...)
	}

	return resources, nil
}

func (r *iamRepository) ListAllUserPolicyAttachments(users []*iam.User) ([]*AttachedUserPolicy, error) {
	var resources []*AttachedUserPolicy
	for _, user := range users {
		cacheKey := fmt.Sprintf("iamListAllUserPolicyAttachments_user_%s", *user.UserName)
		if v := r.cache.Get(cacheKey); v != nil {
			resources = append(resources, v.([]*AttachedUserPolicy)...)
			continue
		}

		userResources := make([]*AttachedUserPolicy, 0)
		input := &iam.ListAttachedUserPoliciesInput{
			UserName: user.UserName,
		}
		err := r.client.ListAttachedUserPoliciesPages(input, func(res *iam.ListAttachedUserPoliciesOutput, lastPage bool) bool {
			for _, policy := range res.AttachedPolicies {
				p := *policy
				userResources = append(userResources, &AttachedUserPolicy{
					AttachedPolicy: p,
					UserName:       *input.UserName,
				})
			}
			return !lastPage
		})
		if err != nil {
			return nil, err
		}

		r.cache.Put(cacheKey, userResources)
		resources = append(resources, userResources...)
	}

	return resources, nil
}

func (r *iamRepository) ListAllUserPolicies(users []*iam.User) ([]string, error) {
	var resources []string
	for _, user := range users {
		cacheKey := fmt.Sprintf("iamListAllUserPolicies_user_%s", *user.UserName)
		if v := r.cache.Get(cacheKey); v != nil {
			resources = append(resources, v.([]string)...)
			continue
		}

		userResources := make([]string, 0)
		input := &iam.ListUserPoliciesInput{
			UserName: user.UserName,
		}
		err := r.client.ListUserPoliciesPages(input, func(res *iam.ListUserPoliciesOutput, lastPage bool) bool {
			for _, polName := range res.PolicyNames {
				userResources = append(userResources, fmt.Sprintf("%s:%s", *input.UserName, *polName))
			}
			return !lastPage
		})
		if err != nil {
			return nil, err
		}

		r.cache.Put(cacheKey, userResources)
		resources = append(resources, userResources...)
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

type RolePolicy struct {
	Policy   string
	RoleName string
}
