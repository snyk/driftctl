package repository

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/pkg/remote/cache"
	awstest "github.com/snyk/driftctl/test/aws"
	"github.com/stretchr/testify/mock"

	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
)

func Test_ELBv2Repository_ListAllLoadBalancers(t *testing.T) {
	dummyError := errors.New("dummy error")

	tests := []struct {
		name    string
		mocks   func(*awstest.MockFakeELBV2, *cache.MockCache)
		want    []*elbv2.LoadBalancer
		wantErr error
	}{
		{
			name: "list load balancers",
			mocks: func(client *awstest.MockFakeELBV2, store *cache.MockCache) {
				results := &elbv2.DescribeLoadBalancersOutput{
					LoadBalancers: []*elbv2.LoadBalancer{
						{
							LoadBalancerArn:  aws.String("test-1"),
							LoadBalancerName: aws.String("test-1"),
						},
						{
							LoadBalancerArn:  aws.String("test-2"),
							LoadBalancerName: aws.String("test-2"),
						},
					},
				}

				store.On("Get", "elbListAllLoadBalancers").Return(nil).Once()

				client.On("DescribeLoadBalancersPages",
					&elbv2.DescribeLoadBalancersInput{},
					mock.MatchedBy(func(callback func(res *elbv2.DescribeLoadBalancersOutput, lastPage bool) bool) bool {
						callback(&elbv2.DescribeLoadBalancersOutput{LoadBalancers: []*elbv2.LoadBalancer{
							results.LoadBalancers[0],
						}}, false)
						callback(&elbv2.DescribeLoadBalancersOutput{LoadBalancers: []*elbv2.LoadBalancer{
							results.LoadBalancers[1],
						}}, true)
						return true
					})).Return(nil).Once()

				store.On("Put", "elbListAllLoadBalancers", results.LoadBalancers).Return(false).Once()
			},
			want: []*elbv2.LoadBalancer{
				{
					LoadBalancerArn:  aws.String("test-1"),
					LoadBalancerName: aws.String("test-1"),
				},
				{
					LoadBalancerArn:  aws.String("test-2"),
					LoadBalancerName: aws.String("test-2"),
				},
			},
		},
		{
			name: "list load balancers from cache",
			mocks: func(client *awstest.MockFakeELBV2, store *cache.MockCache) {
				output := &elbv2.DescribeLoadBalancersOutput{
					LoadBalancers: []*elbv2.LoadBalancer{
						{
							LoadBalancerArn:  aws.String("test-1"),
							LoadBalancerName: aws.String("test-1"),
						},
					},
				}

				store.On("Get", "elbListAllLoadBalancers").Return(output.LoadBalancers).Once()
			},
			want: []*elbv2.LoadBalancer{
				{
					LoadBalancerArn:  aws.String("test-1"),
					LoadBalancerName: aws.String("test-1"),
				},
			},
		},
		{
			name: "error listing load balancers",
			mocks: func(client *awstest.MockFakeELBV2, store *cache.MockCache) {
				store.On("Get", "elbListAllLoadBalancers").Return(nil).Once()

				client.On("DescribeLoadBalancersPages",
					&elbv2.DescribeLoadBalancersInput{},
					mock.MatchedBy(func(callback func(res *elbv2.DescribeLoadBalancersOutput, lastPage bool) bool) bool {
						callback(&elbv2.DescribeLoadBalancersOutput{LoadBalancers: []*elbv2.LoadBalancer{}}, true)
						return true
					})).Return(dummyError).Once()
			},
			wantErr: dummyError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeELBV2{}
			tt.mocks(client, store)
			r := &elbv2Repository{
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
		})
	}
}
