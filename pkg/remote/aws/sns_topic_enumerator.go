package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
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

	results := make([]*resource.Resource, len(topics))

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
