package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type SNSTopicEnumerator struct {
	repository repository.SNSRepository
	factory    resource.ResourceFactory
}

func NewSNSTopicEnumerator(repo repository.SNSRepository, factory resource.ResourceFactory) *SNSTopicEnumerator {
	return &SNSTopicEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *SNSTopicEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsSnsTopicResourceType
}

func (e *SNSTopicEnumerator) Enumerate() ([]*resource.Resource, error) {
	topics, err := e.repository.ListAllTopics()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(topics))

	for _, topic := range topics {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*topic.TopicArn,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
