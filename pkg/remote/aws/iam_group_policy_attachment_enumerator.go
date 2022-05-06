package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type IamGroupPolicyAttachmentEnumerator struct {
	repository repository.IAMRepository
	factory    resource.ResourceFactory
}

func NewIamGroupPolicyAttachmentEnumerator(repo repository.IAMRepository, factory resource.ResourceFactory) *IamGroupPolicyAttachmentEnumerator {
	return &IamGroupPolicyAttachmentEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *IamGroupPolicyAttachmentEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsIamGroupPolicyAttachmentResourceType
}

func (e *IamGroupPolicyAttachmentEnumerator) Enumerate() ([]*resource.Resource, error) {
	groupPolicies, err := e.repository.ListAllPolicies()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(groupPolicies))

	for _, groupPolicy := range groupPolicies {
		if !*groupPolicy.IsAttachable || *groupPolicy.AttachmentCount == 0 {
			continue
		}
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*groupPolicy.PolicyName,
				map[string]interface{}{
					"policy_arn": *groupPolicy.Arn,
				},
			),
		)
	}

	return results, err
}
