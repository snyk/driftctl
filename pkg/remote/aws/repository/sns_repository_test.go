package repository

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	awstest "github.com/cloudskiff/driftctl/test/aws"

	"github.com/stretchr/testify/mock"

	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/sns"
)

func Test_snsRepository_ListAllTopics(t *testing.T) {

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeSNS)
		want    []*sns.Topic
		wantErr error
	}{
		{
			name: "List with 2 pages",
			mocks: func(client *awstest.MockFakeSNS) {
				client.On("ListTopicsPages",
					&sns.ListTopicsInput{},
					mock.MatchedBy(func(callback func(res *sns.ListTopicsOutput, lastPage bool) bool) bool {
						callback(&sns.ListTopicsOutput{
							Topics: []*sns.Topic{
								{TopicArn: aws.String("arn1")},
								{TopicArn: aws.String("arn2")},
								{TopicArn: aws.String("arn3")},
							},
						}, false)
						callback(&sns.ListTopicsOutput{
							Topics: []*sns.Topic{
								{TopicArn: aws.String("arn4")},
								{TopicArn: aws.String("arn5")},
								{TopicArn: aws.String("arn6")},
							},
						}, true)
						return true
					})).Return(nil)
			},
			want: []*sns.Topic{
				{TopicArn: aws.String("arn1")},
				{TopicArn: aws.String("arn2")},
				{TopicArn: aws.String("arn3")},
				{TopicArn: aws.String("arn4")},
				{TopicArn: aws.String("arn5")},
				{TopicArn: aws.String("arn6")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &awstest.MockFakeSNS{}
			tt.mocks(client)
			r := &snsRepository{
				client: client,
			}
			got, err := r.ListAllTopics()
			assert.Equal(t, tt.wantErr, err)
			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

func Test_snsRepository_ListAllSubscriptions(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeSNS)
		want    []*sns.Subscription
		wantErr error
	}{
		{
			name: "List with 2 pages",
			mocks: func(client *awstest.MockFakeSNS) {
				client.On("ListSubscriptionsPages",
					&sns.ListSubscriptionsInput{},
					mock.MatchedBy(func(callback func(res *sns.ListSubscriptionsOutput, lastPage bool) bool) bool {
						callback(&sns.ListSubscriptionsOutput{
							Subscriptions: []*sns.Subscription{
								{TopicArn: aws.String("arn1"), SubscriptionArn: aws.String("SubArn1")},
								{TopicArn: aws.String("arn2"), SubscriptionArn: aws.String("SubArn2")},
								{TopicArn: aws.String("arn3"), SubscriptionArn: aws.String("SubArn3")},
							},
						}, false)
						callback(&sns.ListSubscriptionsOutput{
							Subscriptions: []*sns.Subscription{
								{TopicArn: aws.String("arn4"), SubscriptionArn: aws.String("SubArn4")},
								{TopicArn: aws.String("arn5"), SubscriptionArn: aws.String("SubArn5")},
								{TopicArn: aws.String("arn6"), SubscriptionArn: aws.String("SubArn6")},
							},
						}, true)
						return true
					})).Return(nil)
			},
			want: []*sns.Subscription{
				{TopicArn: aws.String("arn1"), SubscriptionArn: aws.String("SubArn1")},
				{TopicArn: aws.String("arn2"), SubscriptionArn: aws.String("SubArn2")},
				{TopicArn: aws.String("arn3"), SubscriptionArn: aws.String("SubArn3")},
				{TopicArn: aws.String("arn4"), SubscriptionArn: aws.String("SubArn4")},
				{TopicArn: aws.String("arn5"), SubscriptionArn: aws.String("SubArn5")},
				{TopicArn: aws.String("arn6"), SubscriptionArn: aws.String("SubArn6")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &awstest.MockFakeSNS{}
			tt.mocks(client)
			r := &snsRepository{
				client: client,
			}
			got, err := r.ListAllSubscriptions()
			assert.Equal(t, tt.wantErr, err)
			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}
