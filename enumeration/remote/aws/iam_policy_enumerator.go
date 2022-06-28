package aws

import (
	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type IamPolicyEnumerator struct {
	repository repository.IAMRepository
	factory    resource.ResourceFactory
}

func NewIamPolicyEnumerator(repo repository.IAMRepository, factory resource.ResourceFactory) *IamPolicyEnumerator {
	return &IamPolicyEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *IamPolicyEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsIamPolicyResourceType
}

func (e *IamPolicyEnumerator) Enumerate() ([]*resource.Resource, error) {
	policies, err := e.repository.ListAllPolicies()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(policies))

	for _, policy := range policies {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				awssdk.StringValue(policy.Arn),
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
