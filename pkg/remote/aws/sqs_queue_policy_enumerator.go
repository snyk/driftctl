package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	awssdk "github.com/aws/aws-sdk-go/aws"
)

type SQSQueuePolicyEnumerator struct {
	repository repository.SQSRepository
	factory    resource.ResourceFactory
}

func NewSQSQueuePolicyEnumerator(repo repository.SQSRepository, factory resource.ResourceFactory) *SQSQueuePolicyEnumerator {
	return &SQSQueuePolicyEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *SQSQueuePolicyEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsSqsQueuePolicyResourceType
}

func (e *SQSQueuePolicyEnumerator) Enumerate() ([]resource.Resource, error) {
	queues, err := e.repository.ListAllQueues()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, string(e.SupportedType()), aws.AwsSqsQueueResourceType)
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
