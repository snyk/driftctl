package aws

import (
	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
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

func (e *IamUserEnumerator) Enumerate() ([]resource.Resource, error) {
	users, err := e.repository.ListAllUsers()
	if err != nil {
		return nil, remoteerror.NewResourceScanningError(err, string(e.SupportedType()))
	}

	results := make([]resource.Resource, len(users))

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
