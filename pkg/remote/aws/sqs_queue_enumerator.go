package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	awssdk "github.com/aws/aws-sdk-go/aws"
)

type SqsQueueEnumerator struct {
	repository repository.SQSRepository
	factory    resource.ResourceFactory
}

func NewSqsQueueEnumerator(repo repository.SQSRepository, factory resource.ResourceFactory) *SqsQueueEnumerator {
	return &SqsQueueEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *SqsQueueEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsSqsQueueResourceType
}

func (e *SqsQueueEnumerator) Enumerate() ([]resource.Resource, error) {
	queues, err := e.repository.ListAllQueues()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, string(e.SupportedType()))
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
