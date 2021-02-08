package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
)

type SNSRepository interface {
	ListAllTopics() ([]*sns.Topic, error)
	ListAllSubscriptions() ([]*sns.Subscription, error)
}

type snsRepository struct {
	client snsiface.SNSAPI
}

func NewSNSClient(session *session.Session) *snsRepository {
	return &snsRepository{
		sns.New(session),
	}
}

func (r *snsRepository) ListAllTopics() ([]*sns.Topic, error) {
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

func (r *snsRepository) ListAllSubscriptions() ([]*sns.Subscription, error) {
	var subscriptions []*sns.Subscription
	input := &sns.ListSubscriptionsInput{}
	err := r.client.ListSubscriptionsPages(input, func(res *sns.ListSubscriptionsOutput, lastPage bool) bool {
		subscriptions = append(subscriptions, res.Subscriptions...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return subscriptions, nil
}
