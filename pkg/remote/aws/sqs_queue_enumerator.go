package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

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

func (e *SQSQueueEnumerator) Enumerate() ([]resource.Resource, error) {
	queues, err := e.repository.ListAllQueues()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]resource.Resource, len(queues))

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
