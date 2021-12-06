package repository

import (
	"strings"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/snyk/driftctl/pkg/remote/cache"
	awstest "github.com/snyk/driftctl/test/aws"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_sqsRepository_ListAllQueues(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeSQS)
		want    []*string
		wantErr error
	}{
		{
			name: "list with multiple pages",
			mocks: func(client *awstest.MockFakeSQS) {
				client.On("ListQueuesPages",
					&sqs.ListQueuesInput{},
					mock.MatchedBy(func(callback func(res *sqs.ListQueuesOutput, lastPage bool) bool) bool {
						callback(&sqs.ListQueuesOutput{
							QueueUrls: []*string{
								awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/bar.fifo"),
								awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/foo"),
							},
						}, false)
						callback(&sqs.ListQueuesOutput{
							QueueUrls: []*string{
								awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/baz"),
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			want: []*string{
				awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/bar.fifo"),
				awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/foo"),
				awssdk.String("https://sqs.eu-west-3.amazonaws.com/047081014315/baz"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := &awstest.MockFakeSQS{}
			tt.mocks(client)
			r := &sqsRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllQueues()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllQueues()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*string{}, store.Get("sqsListAllQueues"))
			}

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

func Test_sqsRepository_GetQueueAttributes(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeSQS)
		want    *sqs.GetQueueAttributesOutput
		wantErr error
	}{
		{
			name: "get attributes",
			mocks: func(client *awstest.MockFakeSQS) {
				client.On(
					"GetQueueAttributes",
					&sqs.GetQueueAttributesInput{
						AttributeNames: awssdk.StringSlice([]string{sqs.QueueAttributeNamePolicy}),
						QueueUrl:       awssdk.String("http://example.com"),
					},
				).Return(
					&sqs.GetQueueAttributesOutput{
						Attributes: map[string]*string{
							sqs.QueueAttributeNamePolicy: awssdk.String("foobar"),
						},
					},
					nil,
				).Once()
			},
			want: &sqs.GetQueueAttributesOutput{
				Attributes: map[string]*string{
					sqs.QueueAttributeNamePolicy: awssdk.String("foobar"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := &awstest.MockFakeSQS{}
			tt.mocks(client)
			r := &sqsRepository{
				client: client,
				cache:  store,
			}
			got, err := r.GetQueueAttributes("http://example.com")
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.GetQueueAttributes("http://example.com")
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, &sqs.GetQueueAttributesOutput{}, store.Get("sqsGetQueueAttributes_http://example.com"))
			}

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
