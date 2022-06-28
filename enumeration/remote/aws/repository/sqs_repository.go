package repository

import (
	"fmt"
	"github.com/snyk/driftctl/enumeration/remote/cache"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

type SQSRepository interface {
	ListAllQueues() ([]*string, error)
	GetQueueAttributes(url string) (*sqs.GetQueueAttributesOutput, error)
}

type sqsRepository struct {
	client sqsiface.SQSAPI
	cache  cache.Cache
}

func NewSQSRepository(session *session.Session, c cache.Cache) *sqsRepository {
	return &sqsRepository{
		sqs.New(session),
		c,
	}
}

func (r *sqsRepository) GetQueueAttributes(url string) (*sqs.GetQueueAttributesOutput, error) {
	cacheKey := fmt.Sprintf("sqsGetQueueAttributes_%s", url)
	if v := r.cache.Get(cacheKey); v != nil {
		return v.(*sqs.GetQueueAttributesOutput), nil
	}

	attributes, err := r.client.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		AttributeNames: aws.StringSlice([]string{sqs.QueueAttributeNamePolicy}),
		QueueUrl:       &url,
	})
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, attributes)

	return attributes, nil
}

func (r *sqsRepository) ListAllQueues() ([]*string, error) {

	cacheKey := "sqsListAllQueues"
	v := r.cache.GetAndLock(cacheKey)
	defer r.cache.Unlock(cacheKey)
	if v != nil {
		return v.([]*string), nil
	}

	var queues []*string
	input := sqs.ListQueuesInput{}
	err := r.client.ListQueuesPages(&input,
		func(resp *sqs.ListQueuesOutput, lastPage bool) bool {
			queues = append(queues, resp.QueueUrls...)
			return !lastPage
		},
	)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, queues)
	return queues, nil
}
