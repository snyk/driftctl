package aws

import (
	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type IamUserEnumerator struct {
	repository repository.IAMRepository
	factory    resource.ResourceFactory
}

func NewIamUserEnumerator(repo repository.IAMRepository, factory resource.ResourceFactory) *IamUserEnumerator {
	return &IamUserEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *IamUserEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsIamUserResourceType
}

func (e *IamUserEnumerator) Enumerate() ([]*resource.Resource, error) {
	users, err := e.repository.ListAllUsers()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(users))

	for _, user := range users {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				awssdk.StringValue(user.UserName),
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
