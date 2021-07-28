package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type SNSTopicPolicyEnumerator struct {
	repository repository.SNSRepository
	factory    resource.ResourceFactory
}

func NewSNSTopicPolicyEnumerator(repo repository.SNSRepository, factory resource.ResourceFactory) *SNSTopicPolicyEnumerator {
	return &SNSTopicPolicyEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *SNSTopicPolicyEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsSnsTopicPolicyResourceType
}

func (e *SNSTopicPolicyEnumerator) Enumerate() ([]resource.Resource, error) {
	topics, err := e.repository.ListAllTopics()
	if err != nil {
		return nil, remoteerror.NewResourceScanningErrorWithType(err, string(e.SupportedType()), aws.AwsSnsTopicResourceType)
	}

	results := make([]resource.Resource, len(topics))

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
