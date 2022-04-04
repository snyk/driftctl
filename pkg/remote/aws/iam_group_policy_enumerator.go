package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type IamGroupPolicyEnumerator struct {
	repository repository.IAMRepository
	factory    resource.ResourceFactory
}

func NewIamGroupPolicyEnumerator(repo repository.IAMRepository, factory resource.ResourceFactory) *IamGroupPolicyEnumerator {
	return &IamGroupPolicyEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *IamGroupPolicyEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsIamGroupPolicyResourceType
}

func (e *IamGroupPolicyEnumerator) Enumerate() ([]*resource.Resource, error) {
	groups, err := e.repository.ListAllGroups()
	if err != nil {
		// TODO Use constant instead of string for `aws_iam_group` here when we'll add support for the group resource
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), "aws_iam_group")
	}
	groupPolicies, err := e.repository.ListAllGroupPolicies(groups)
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(groupPolicies))

	for _, groupPolicy := range groupPolicies {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				groupPolicy,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
