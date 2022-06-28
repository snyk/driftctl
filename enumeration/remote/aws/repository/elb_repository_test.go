package repository

import (
	"errors"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	awstest "github.com/snyk/driftctl/test/aws"

	"github.com/stretchr/testify/mock"

	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
)

func Test_ELBRepository_ListAllLoadBalancers(t *testing.T) {
	dummyErr := errors.New("dummy error")

	results := []*elb.LoadBalancerDescription{
		{
			LoadBalancerName: aws.String("test-lb-1"),
		},
		{
			LoadBalancerName: aws.String("test-lb-2"),
		},
	}

	tests := []struct {
		name    string
		mocks   func(*awstest.MockFakeELB, *cache.MockCache)
		want    []*elb.LoadBalancerDescription
		wantErr error
	}{
		{
			name: "List load balancers with multiple pages",
			mocks: func(client *awstest.MockFakeELB, store *cache.MockCache) {
				store.On("Get", "elbListAllLoadBalancers").Return(nil).Once()

				client.On("DescribeLoadBalancersPages",
					&elb.DescribeLoadBalancersInput{},
					mock.MatchedBy(func(callback func(res *elb.DescribeLoadBalancersOutput, lastPage bool) bool) bool {
						callback(&elb.DescribeLoadBalancersOutput{LoadBalancerDescriptions: []*elb.LoadBalancerDescription{
							results[0],
						}}, false)
						callback(&elb.DescribeLoadBalancersOutput{LoadBalancerDescriptions: []*elb.LoadBalancerDescription{
							results[1],
						}}, true)
						return true
					})).Return(nil).Once()

				store.On("Put", "elbListAllLoadBalancers", results).Return(false).Once()
			},
			want: []*elb.LoadBalancerDescription{
				{
					LoadBalancerName: aws.String("test-lb-1"),
				},
				{
					LoadBalancerName: aws.String("test-lb-2"),
				},
			},
		},
		{
			name: "List load balancers with multiple pages (cache hit)",
			mocks: func(client *awstest.MockFakeELB, store *cache.MockCache) {
				store.On("Get", "elbListAllLoadBalancers").Return(results).Once()
			},
			want: []*elb.LoadBalancerDescription{
				{
					LoadBalancerName: aws.String("test-lb-1"),
				},
				{
					LoadBalancerName: aws.String("test-lb-2"),
				},
			},
		},
		{
			name: "Error listing load balancers",
			mocks: func(client *awstest.MockFakeELB, store *cache.MockCache) {
				store.On("Get", "elbListAllLoadBalancers").Return(nil).Once()

				client.On("DescribeLoadBalancersPages",
					&elb.DescribeLoadBalancersInput{},
					mock.MatchedBy(func(callback func(res *elb.DescribeLoadBalancersOutput, lastPage bool) bool) bool {
						callback(&elb.DescribeLoadBalancersOutput{LoadBalancerDescriptions: []*elb.LoadBalancerDescription{}}, true)
						return true
					})).Return(dummyErr).Once()
			},
			wantErr: dummyErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeELB{}
			tt.mocks(client, store)
			r := &elbRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllLoadBalancers()
			assert.Equal(t, tt.wantErr, err)

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %v -> %v", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}

			client.AssertExpectations(t)
			store.AssertExpectations(t)
		})
	}
}
