package repository

import (
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	awstest "github.com/cloudskiff/driftctl/test/aws"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
)

func Test_AutoscalingRepository_DescribeLaunchConfigurations(t *testing.T) {
	dummryError := errors.New("dummy error")

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeAutoscaling)
		want    []*autoscaling.LaunchConfiguration
		wantErr error
	}{
		{
			name: "List all launch configurations",
			mocks: func(client *awstest.MockFakeAutoscaling) {
				client.On("DescribeLaunchConfigurations",
					&autoscaling.DescribeLaunchConfigurationsInput{}).Return(&autoscaling.DescribeLaunchConfigurationsOutput{
					LaunchConfigurations: []*autoscaling.LaunchConfiguration{
						{ImageId: aws.String("1")},
						{ImageId: aws.String("2")},
						{ImageId: aws.String("3")},
						{ImageId: aws.String("4")},
					},
				}, nil).Once()
			},
			want: []*autoscaling.LaunchConfiguration{
				{ImageId: aws.String("1")},
				{ImageId: aws.String("2")},
				{ImageId: aws.String("3")},
				{ImageId: aws.String("4")},
			},
		},
		{
			name: "Error listing all launch configurations",
			mocks: func(client *awstest.MockFakeAutoscaling) {
				client.On("DescribeLaunchConfigurations",
					&autoscaling.DescribeLaunchConfigurationsInput{}).Return(nil, dummryError).Once()
			},
			want:    nil,
			wantErr: dummryError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := &awstest.MockFakeAutoscaling{}
			tt.mocks(client)
			r := &autoScalingRepository{
				client: client,
				cache:  store,
			}
			got, err := r.DescribeLaunchConfigurations()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.DescribeLaunchConfigurations()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*autoscaling.LaunchConfiguration{}, store.Get("DescribeLaunchConfigurations"))
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
