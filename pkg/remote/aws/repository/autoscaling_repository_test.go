package repository

import (
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	awstest "github.com/cloudskiff/driftctl/test/aws"
	"github.com/stretchr/testify/mock"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
)

func Test_AutoscalingRepository_DescribeLaunchConfigurations(t *testing.T) {
	dummryError := errors.New("dummy error")

	expectedLaunchConfigurations := []*autoscaling.LaunchConfiguration{
		{ImageId: aws.String("1")},
		{ImageId: aws.String("2")},
		{ImageId: aws.String("3")},
		{ImageId: aws.String("4")},
	}

	tests := []struct {
		name    string
		mocks   func(*awstest.MockFakeAutoscaling, *cache.MockCache)
		want    []*autoscaling.LaunchConfiguration
		wantErr error
	}{
		{
			name: "List all launch configurations",
			mocks: func(client *awstest.MockFakeAutoscaling, store *cache.MockCache) {
				store.On("Get", "DescribeLaunchConfigurations").Return(nil).Once()

				client.On("DescribeLaunchConfigurationsPages",
					&autoscaling.DescribeLaunchConfigurationsInput{},
					mock.MatchedBy(func(callback func(res *autoscaling.DescribeLaunchConfigurationsOutput, lastPage bool) bool) bool {
						callback(&autoscaling.DescribeLaunchConfigurationsOutput{
							LaunchConfigurations: expectedLaunchConfigurations[:2],
						}, false)
						callback(&autoscaling.DescribeLaunchConfigurationsOutput{
							LaunchConfigurations: expectedLaunchConfigurations[2:],
						}, true)
						return true
					})).Return(nil).Once()

				store.On("Put", "DescribeLaunchConfigurations", expectedLaunchConfigurations).Return(false).Once()
			},
			want: expectedLaunchConfigurations,
		},
		{
			name: "Hit cache and list all launch configurations",
			mocks: func(client *awstest.MockFakeAutoscaling, store *cache.MockCache) {
				store.On("Get", "DescribeLaunchConfigurations").Return(expectedLaunchConfigurations).Once()
			},
			want: expectedLaunchConfigurations,
		},
		{
			name: "Error listing all launch configurations",
			mocks: func(client *awstest.MockFakeAutoscaling, store *cache.MockCache) {
				store.On("Get", "DescribeLaunchConfigurations").Return(nil).Once()

				client.On("DescribeLaunchConfigurationsPages", &autoscaling.DescribeLaunchConfigurationsInput{}, mock.MatchedBy(func(callback func(res *autoscaling.DescribeLaunchConfigurationsOutput, lastPage bool) bool) bool {
					callback(&autoscaling.DescribeLaunchConfigurationsOutput{
						LaunchConfigurations: []*autoscaling.LaunchConfiguration{},
					}, true)
					return true
				})).Return(dummryError).Once()
			},
			want:    nil,
			wantErr: dummryError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeAutoscaling{}
			tt.mocks(client, store)
			r := &autoScalingRepository{
				client: client,
				cache:  store,
			}
			got, err := r.DescribeLaunchConfigurations()
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
