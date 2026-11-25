package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"

	awssdk "github.com/aws/aws-sdk-go/aws"
)

type SQSQueueEnumerator struct {
	repository repository.SQSRepository
	factory    resource.ResourceFactory
}

func NewSQSQueueEnumerator(repo repository.SQSRepository, factory resource.ResourceFactory) *SQSQueueEnumerator {
	return &SQSQueueEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *SQSQueueEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsSqsQueueResourceType
}

func (e *SQSQueueEnumerator) Enumerate() ([]*resource.Resource, error) {
	queues, err := e.repository.ListAllQueues()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(queues))

	for _, queue := range queues {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				awssdk.StringValue(queue),
				map[string]interface{}{},
			),
		)
	}
	return results, err
}
