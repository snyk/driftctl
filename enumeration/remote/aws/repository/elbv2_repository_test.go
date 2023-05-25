package repository

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/pkg/errors"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	awstest "github.com/snyk/driftctl/test/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_ELBV2Repository_ListAllLoadBalancers(t *testing.T) {
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

				store.On("GetAndLock", "elbv2ListAllLoadBalancers").Return(nil).Once()
				store.On("Unlock", "elbv2ListAllLoadBalancers").Return().Once()

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

				store.On("Put", "elbv2ListAllLoadBalancers", results.LoadBalancers).Return(false).Once()
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

				store.On("GetAndLock", "elbv2ListAllLoadBalancers").Return(output.LoadBalancers).Once()
				store.On("Unlock", "elbv2ListAllLoadBalancers").Return().Once()
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
				store.On("GetAndLock", "elbv2ListAllLoadBalancers").Return(nil).Once()
				store.On("Unlock", "elbv2ListAllLoadBalancers").Return().Once()

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

func Test_ELBV2Repository_ListAllLoadBalancerListeners(t *testing.T) {
	dummyError := errors.New("dummy error")

	type call struct {
		loadBalancerArn string
		mocks           func(*awstest.MockFakeELBV2, *cache.MockCache)
		want            []*elbv2.Listener
		wantErr         error
	}

	tests := []struct {
		name  string
		calls []call
	}{
		{
			name: "list load balancer listeners",
			calls: []call{
				{
					loadBalancerArn: "test-lb",
					mocks: func(client *awstest.MockFakeELBV2, store *cache.MockCache) {
						results := &elbv2.DescribeListenersOutput{
							Listeners: []*elbv2.Listener{
								{
									LoadBalancerArn: aws.String("test-lb"),
									ListenerArn:     aws.String("test-lb-listener-1"),
								},
								{
									LoadBalancerArn: aws.String("test-lb"),
									ListenerArn:     aws.String("test-lb-listener-2"),
								},
							},
						}

						store.On("Get", "elbv2ListAllLoadBalancerListeners_test-lb").Return(nil).Once()

						client.On("DescribeListenersPages",
							&elbv2.DescribeListenersInput{LoadBalancerArn: aws.String("test-lb")},
							mock.MatchedBy(func(callback func(res *elbv2.DescribeListenersOutput, lastPage bool) bool) bool {
								callback(&elbv2.DescribeListenersOutput{Listeners: []*elbv2.Listener{
									results.Listeners[0],
								}}, false)
								callback(&elbv2.DescribeListenersOutput{Listeners: []*elbv2.Listener{
									results.Listeners[1],
								}}, true)
								return true
							})).Return(nil).Once()

						store.On("Put", "elbv2ListAllLoadBalancerListeners_test-lb", results.Listeners).Return(false).Once()
					},
					want: []*elbv2.Listener{
						{
							LoadBalancerArn: aws.String("test-lb"),
							ListenerArn:     aws.String("test-lb-listener-1"),
						},
						{
							LoadBalancerArn: aws.String("test-lb"),
							ListenerArn:     aws.String("test-lb-listener-2"),
						},
					},
				},
			},
		},
		{
			name: "list load balancer listeners from cache",
			calls: []call{
				{
					loadBalancerArn: "test-lb",
					mocks: func(client *awstest.MockFakeELBV2, store *cache.MockCache) {
						output := &elbv2.DescribeListenersOutput{
							Listeners: []*elbv2.Listener{
								{
									LoadBalancerArn: aws.String("test-lb"),
									ListenerArn:     aws.String("test-lb-listener"),
								},
							},
						}

						store.On("Get", "elbv2ListAllLoadBalancerListeners_test-lb").Return(output.Listeners).Once()
					},
					want: []*elbv2.Listener{
						{
							LoadBalancerArn: aws.String("test-lb"),
							ListenerArn:     aws.String("test-lb-listener"),
						},
					},
				},
			},
		},
		{
			name: "list load balancer listeners from multiple load balancers",
			calls: []call{
				{
					loadBalancerArn: "test-lb-1",
					mocks: func(client *awstest.MockFakeELBV2, store *cache.MockCache) {
						output := &elbv2.DescribeListenersOutput{
							Listeners: []*elbv2.Listener{
								{
									LoadBalancerArn: aws.String("test-lb-1"),
									ListenerArn:     aws.String("test-lb-1-listener"),
								},
							},
						}

						store.On("Get", "elbv2ListAllLoadBalancerListeners_test-lb-1").Return(output.Listeners).Once()
					},
					want: []*elbv2.Listener{
						{
							LoadBalancerArn: aws.String("test-lb-1"),
							ListenerArn:     aws.String("test-lb-1-listener"),
						},
					},
				},
				{
					loadBalancerArn: "test-lb-2",
					mocks: func(client *awstest.MockFakeELBV2, store *cache.MockCache) {
						output := &elbv2.DescribeListenersOutput{
							Listeners: []*elbv2.Listener{
								{
									LoadBalancerArn: aws.String("test-lb-2"),
									ListenerArn:     aws.String("test-lb-2-listener"),
								},
							},
						}

						store.On("Get", "elbv2ListAllLoadBalancerListeners_test-lb-2").Return(output.Listeners).Once()
					},
					want: []*elbv2.Listener{
						{
							LoadBalancerArn: aws.String("test-lb-2"),
							ListenerArn:     aws.String("test-lb-2-listener"),
						},
					},
				},
			},
		},
		{
			name: "error listing load balancer listeners",
			calls: []call{
				{
					loadBalancerArn: "test-lb",
					mocks: func(client *awstest.MockFakeELBV2, store *cache.MockCache) {
						store.On("Get", "elbv2ListAllLoadBalancerListeners_test-lb").Return(nil).Once()

						client.On("DescribeListenersPages",
							&elbv2.DescribeListenersInput{LoadBalancerArn: aws.String("test-lb")},
							mock.MatchedBy(func(callback func(res *elbv2.DescribeListenersOutput, lastPage bool) bool) bool {
								callback(&elbv2.DescribeListenersOutput{Listeners: []*elbv2.Listener{}}, true)
								return true
							})).Return(dummyError).Once()
					},
					wantErr: dummyError,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeELBV2{}

			for _, call := range tt.calls {
				call.mocks(client, store)
				r := &elbv2Repository{
					client: client,
					cache:  store,
				}
				got, err := r.ListAllLoadBalancerListeners(call.loadBalancerArn)
				assert.Equal(t, call.wantErr, err)

				changelog, err := diff.Diff(got, call.want)
				assert.Nil(t, err)
				if len(changelog) > 0 {
					for _, change := range changelog {
						t.Errorf("%s: %v -> %v", strings.Join(change.Path, "."), change.From, change.To)
					}
					t.Fail()
				}
			}
		})
	}
}
