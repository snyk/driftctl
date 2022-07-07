package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type IamGroupEnumerator struct {
	repository repository.IAMRepository
	factory    resource.ResourceFactory
}

func NewIamGroupEnumerator(repo repository.IAMRepository, factory resource.ResourceFactory) *IamGroupEnumerator {
	return &IamGroupEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *IamGroupEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsIamGroupResourceType
}

func (e *IamGroupEnumerator) Enumerate() ([]*resource.Resource, error) {
	groups, err := e.repository.ListAllGroups()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsIamGroupResourceType)
	}

	results := make([]*resource.Resource, 0, len(groups))

	for _, group := range groups {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*group.GroupName,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
