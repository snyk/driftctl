package repository

import (
	"strings"
	"testing"

	"github.com/snyk/driftctl/enumeration/remote/cache"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	awstest "github.com/snyk/driftctl/test/aws"

	"github.com/stretchr/testify/mock"

	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
)

func Test_cloudtrailRepository_ListAllTrails(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeCloudtrail)
		want    []*cloudtrail.TrailInfo
		wantErr error
	}{
		{
			name: "list multiple trail",
			mocks: func(client *awstest.MockFakeCloudtrail) {
				client.On("ListTrailsPages",
					&cloudtrail.ListTrailsInput{},
					mock.MatchedBy(func(callback func(res *cloudtrail.ListTrailsOutput, lastPage bool) bool) bool {
						callback(&cloudtrail.ListTrailsOutput{
							Trails: []*cloudtrail.TrailInfo{
								{Name: aws.String("trail1")},
								{Name: aws.String("trail2")},
								{Name: aws.String("trail3")},
							},
						}, false)
						callback(&cloudtrail.ListTrailsOutput{
							Trails: []*cloudtrail.TrailInfo{
								{Name: aws.String("trail4")},
								{Name: aws.String("trail5")},
								{Name: aws.String("trail6")},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			want: []*cloudtrail.TrailInfo{
				{Name: aws.String("trail1")},
				{Name: aws.String("trail2")},
				{Name: aws.String("trail3")},
				{Name: aws.String("trail4")},
				{Name: aws.String("trail5")},
				{Name: aws.String("trail6")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := awstest.MockFakeCloudtrail{}
			tt.mocks(&client)
			r := &cloudtrailRepository{
				client: &client,
				cache:  store,
			}
			got, err := r.ListAllTrails()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllTrails()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*cloudtrail.TrailInfo{}, store.Get("ListAllTrails"))
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
