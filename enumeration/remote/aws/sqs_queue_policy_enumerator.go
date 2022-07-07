package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"strings"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"

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

func (e *SQSQueuePolicyEnumerator) Enumerate() ([]*resource.Resource, error) {
	queues, err := e.repository.ListAllQueues()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsSqsQueueResourceType)
	}

	results := make([]*resource.Resource, 0, len(queues))

	for _, queue := range queues {
		attrs := map[string]interface{}{
			"policy": "",
		}
		attributes, err := e.repository.GetQueueAttributes(*queue)
		if err != nil {
			if strings.Contains(err.Error(), "NonExistentQueue") {
				logrus.WithFields(logrus.Fields{
					"queue": *queue,
					"type":  aws.AwsSqsQueueResourceType,
				}).Debugf("Ignoring queue that seems to be already deleted: %+v", err)
				continue
			}
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
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
