package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

type SQSRepository interface {
	ListAllQueues() ([]*string, error)
}

type sqsRepository struct {
	client sqsiface.SQSAPI
}

func NewSQSClient(session *session.Session) *sqsRepository {
	return &sqsRepository{
		sqs.New(session),
	}
}

func (r *sqsRepository) ListAllQueues() ([]*string, error) {
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
	return queues, nil
}
