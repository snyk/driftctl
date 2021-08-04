package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type IamUserPolicyEnumerator struct {
	repository repository.IAMRepository
	factory    resource.ResourceFactory
}

func NewIamUserPolicyEnumerator(repo repository.IAMRepository, factory resource.ResourceFactory) *IamUserPolicyEnumerator {
	return &IamUserPolicyEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *IamUserPolicyEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsIamUserPolicyResourceType
}

func (e *IamUserPolicyEnumerator) Enumerate() ([]resource.Resource, error) {
	users, err := e.repository.ListAllUsers()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsIamUserResourceType)
	}
	userPolicies, err := e.repository.ListAllUserPolicies(users)
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]resource.Resource, len(userPolicies))

	for _, userPolicy := range userPolicies {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				userPolicy,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
