package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
)

type SNSRepository interface {
	ListAllTopics() ([]*sns.Topic, error)
}

type snsRepositoryImpl struct {
	client snsiface.SNSAPI
}

func NewSNSClient(session *session.Session) *snsRepositoryImpl {
	return &snsRepositoryImpl{
		sns.New(session),
	}
}

func (r *snsRepositoryImpl) ListAllTopics() ([]*sns.Topic, error) {
	var topics []*sns.Topic
	input := &sns.ListTopicsInput{}
	err := r.client.ListTopicsPages(input, func(res *sns.ListTopicsOutput, lastPage bool) bool {
		topics = append(topics, res.Topics...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return topics, nil
}
