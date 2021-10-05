package repository

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	awstest "github.com/cloudskiff/driftctl/test/aws"
	"github.com/pkg/errors"

	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
)

func Test_appautoscalingRepository_DescribeScalableTargets(t *testing.T) {
	type args struct {
		namespace string
	}

	tests := []struct {
		name    string
		args    args
		mocks   func(*awstest.MockFakeApplicationAutoScaling, *cache.MockCache)
		want    []*applicationautoscaling.ScalableTarget
		wantErr error
	}{
		{
			name: "should return remote error",
			args: args{
				namespace: "test",
			},
			mocks: func(client *awstest.MockFakeApplicationAutoScaling, c *cache.MockCache) {
				client.On("DescribeScalableTargets",
					&applicationautoscaling.DescribeScalableTargetsInput{
						ServiceNamespace: aws.String("test"),
					}).Return(nil, errors.New("remote error")).Once()

				c.On("Get", "appAutoScalingDescribeScalableTargets_test").Return(nil).Once()
			},
			want:    nil,
			wantErr: errors.New("remote error"),
		},
		{
			name: "should return scalable targets",
			args: args{
				namespace: "test",
			},
			mocks: func(client *awstest.MockFakeApplicationAutoScaling, c *cache.MockCache) {
				results := []*applicationautoscaling.ScalableTarget{
					{
						RoleARN: aws.String("test_target"),
					},
				}

				client.On("DescribeScalableTargets",
					&applicationautoscaling.DescribeScalableTargetsInput{
						ServiceNamespace: aws.String("test"),
					}).Return(&applicationautoscaling.DescribeScalableTargetsOutput{
					ScalableTargets: results,
				}, nil).Once()

				c.On("Get", "appAutoScalingDescribeScalableTargets_test").Return(nil).Once()
				c.On("Put", "appAutoScalingDescribeScalableTargets_test", results).Return(true).Once()
			},
			want: []*applicationautoscaling.ScalableTarget{
				{
					RoleARN: aws.String("test_target"),
				},
			},
		},
		{
			name: "should hit cache return scalable targets",
			args: args{
				namespace: "test",
			},
			mocks: func(client *awstest.MockFakeApplicationAutoScaling, c *cache.MockCache) {
				results := []*applicationautoscaling.ScalableTarget{
					{
						RoleARN: aws.String("test_target"),
					},
				}

				c.On("Get", "appAutoScalingDescribeScalableTargets_test").Return(results).Once()
			},
			want: []*applicationautoscaling.ScalableTarget{
				{
					RoleARN: aws.String("test_target"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeApplicationAutoScaling{}
			tt.mocks(client, store)

			r := &appAutoScalingRepository{
				client: client,
				cache:  store,
			}
			got, err := r.DescribeScalableTargets(tt.args.namespace)
			if err != nil {
				assert.EqualError(t, tt.wantErr, err.Error())
			} else {
				assert.Equal(t, tt.wantErr, err)
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}

			client.AssertExpectations(t)
			store.AssertExpectations(t)
		})
	}
}

func Test_appautoscalingRepository_DescribeScalingPolicies(t *testing.T) {
	type args struct {
		namespace string
	}

	tests := []struct {
		name    string
		args    args
		mocks   func(*awstest.MockFakeApplicationAutoScaling, *cache.MockCache)
		want    []*applicationautoscaling.ScalingPolicy
		wantErr error
	}{
		{
			name: "should return remote error",
			args: args{
				namespace: "test",
			},
			mocks: func(client *awstest.MockFakeApplicationAutoScaling, c *cache.MockCache) {
				client.On("DescribeScalingPolicies",
					&applicationautoscaling.DescribeScalingPoliciesInput{
						ServiceNamespace: aws.String("test"),
					}).Return(nil, errors.New("remote error")).Once()

				c.On("Get", "appAutoScalingDescribeScalingPolicies_test").Return(nil).Once()
			},
			want:    nil,
			wantErr: errors.New("remote error"),
		},
		{
			name: "should return scaling policies",
			args: args{
				namespace: "test",
			},
			mocks: func(client *awstest.MockFakeApplicationAutoScaling, c *cache.MockCache) {
				results := []*applicationautoscaling.ScalingPolicy{
					{
						PolicyARN: aws.String("test_policy"),
					},
				}

				client.On("DescribeScalingPolicies",
					&applicationautoscaling.DescribeScalingPoliciesInput{
						ServiceNamespace: aws.String("test"),
					}).Return(&applicationautoscaling.DescribeScalingPoliciesOutput{
					ScalingPolicies: results,
				}, nil).Once()

				c.On("Get", "appAutoScalingDescribeScalingPolicies_test").Return(nil).Once()
				c.On("Put", "appAutoScalingDescribeScalingPolicies_test", results).Return(true).Once()
			},
			want: []*applicationautoscaling.ScalingPolicy{
				{
					PolicyARN: aws.String("test_policy"),
				},
			},
		},
		{
			name: "should hit cache return scaling policies",
			args: args{
				namespace: "test",
			},
			mocks: func(client *awstest.MockFakeApplicationAutoScaling, c *cache.MockCache) {
				results := []*applicationautoscaling.ScalingPolicy{
					{
						PolicyARN: aws.String("test_policy"),
					},
				}

				c.On("Get", "appAutoScalingDescribeScalingPolicies_test").Return(results).Once()
			},
			want: []*applicationautoscaling.ScalingPolicy{
				{
					PolicyARN: aws.String("test_policy"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeApplicationAutoScaling{}
			tt.mocks(client, store)

			r := &appAutoScalingRepository{
				client: client,
				cache:  store,
			}
			got, err := r.DescribeScalingPolicies(tt.args.namespace)
			if err != nil {
				assert.EqualError(t, tt.wantErr, err.Error())
			} else {
				assert.Equal(t, tt.wantErr, err)
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}

			client.AssertExpectations(t)
			store.AssertExpectations(t)
		})
	}
}
