package client

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
)

type SNSClient interface {
	snsiface.SNSAPI
	ListAllTopics() ([]*sns.Topic, error)
}

type SNSClientImpl struct {
	snsiface.SNSAPI
}

func NewSNSClient(session *session.Session) *SNSClientImpl {
	return &SNSClientImpl{
		sns.New(session),
	}
}

func (c *SNSClientImpl) ListAllTopics() ([]*sns.Topic, error) {
	var topics []*sns.Topic
	input := &sns.ListTopicsInput{}
	err := c.ListTopicsPages(input, func(res *sns.ListTopicsOutput, lastPage bool) bool {
		topics = append(topics, res.Topics...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return topics, nil
}
