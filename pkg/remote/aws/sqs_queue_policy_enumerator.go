package aws

import (
	"github.com/aws/aws-sdk-go/service/sqs"
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

	results := make([]resource.Resource, 0, len(queues))

	for _, queue := range queues {
		attrs := map[string]interface{}{
			"policy": "",
		}
		attributes, err := e.repository.GetQueueAttributes(*queue)
		if err != nil {
			return nil, remoteerror.NewResourceEnumerationError(err, string(e.SupportedType()))
		}
		if attributes.Attributes != nil {
			attrs["policy"] = *attributes.Attributes[sqs.QueueAttributeNamePolicy]
		}
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				awssdk.StringValue(queue),
				attrs,
			),
		)
	}

	return results, err
}
