package repository

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	awstest "github.com/cloudskiff/driftctl/test/aws"

	"github.com/stretchr/testify/mock"

	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
)

func Test_cloudformationRepository_ListAllStacks(t *testing.T) {
	stacks := []*cloudformation.Stack{
		{StackId: aws.String("stack1")},
		{StackId: aws.String("stack2")},
		{StackId: aws.String("stack3")},
		{StackId: aws.String("stack4")},
		{StackId: aws.String("stack5")},
		{StackId: aws.String("stack6")},
	}

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeCloudformation, store *cache.MockCache)
		want    []*cloudformation.Stack
		wantErr error
	}{
		{
			name: "list multiple stacks",
			mocks: func(client *awstest.MockFakeCloudformation, store *cache.MockCache) {
				client.On("DescribeStacksPages",
					&cloudformation.DescribeStacksInput{},
					mock.MatchedBy(func(callback func(res *cloudformation.DescribeStacksOutput, lastPage bool) bool) bool {
						callback(&cloudformation.DescribeStacksOutput{
							Stacks: stacks[:3],
						}, false)
						callback(&cloudformation.DescribeStacksOutput{
							Stacks: stacks[3:],
						}, true)
						return true
					})).Return(nil).Once()

				store.On("Get", "cloudformationListAllStacks").Return(nil).Times(1)
				store.On("Put", "cloudformationListAllStacks", stacks).Return(false).Times(1)
			},
			want: stacks,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeCloudformation, store *cache.MockCache) {
				store.On("Get", "cloudformationListAllStacks").Return(stacks).Times(1)
			},
			want: stacks,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeCloudformation{}
			tt.mocks(client, store)
			r := &cloudformationRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllStacks()
			assert.Equal(t, tt.wantErr, err)

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			store.AssertExpectations(t)
			client.AssertExpectations(t)
		})
	}
}
