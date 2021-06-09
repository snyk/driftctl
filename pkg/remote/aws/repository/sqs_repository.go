package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type SQSRepository interface {
	ListAllQueues() ([]*string, error)
}

type sqsRepository struct {
	client sqsiface.SQSAPI
	cache  cache.Cache
}

func NewSQSClient(session *session.Session, c cache.Cache) *sqsRepository {
	return &sqsRepository{
		sqs.New(session),
		c,
	}
}

func (r *sqsRepository) ListAllQueues() ([]*string, error) {
	if v := r.cache.Get("sqsListAllQueues"); v != nil {
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

	r.cache.Put("sqsListAllQueues", queues)
	return queues, nil
}
