package repository

import (
	"strings"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_sqsRepository_ListAllQueues(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *mocks.FakeSQS)
		want    []*string
		wantErr error
	}{
		{
			name: "list with multiple pages",
			mocks: func(client *mocks.FakeSQS) {
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
					})).Return(nil)
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
			client := &mocks.FakeSQS{}
			tt.mocks(client)
			r := &sqsRepository{
				client: client,
			}
			got, err := r.ListAllQueues()
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
