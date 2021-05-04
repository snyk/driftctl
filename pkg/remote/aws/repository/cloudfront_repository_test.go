package repository

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	awstest "github.com/cloudskiff/driftctl/test/aws"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/stretchr/testify/mock"

	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
)

func Test_cloudfrontRepository_ListAllDistributions(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeCloudFront)
		want    []*cloudfront.DistributionSummary
		wantErr error
	}{
		{
			name: "list multiple distributions",
			mocks: func(client *awstest.MockFakeCloudFront) {
				client.On("ListDistributionsPages",
					&cloudfront.ListDistributionsInput{},
					mock.MatchedBy(func(callback func(res *cloudfront.ListDistributionsOutput, lastPage bool) bool) bool {
						callback(&cloudfront.ListDistributionsOutput{
							DistributionList: &cloudfront.DistributionList{
								Items: []*cloudfront.DistributionSummary{
									{Id: aws.String("distribution1")},
									{Id: aws.String("distribution2")},
									{Id: aws.String("distribution3")},
								},
							},
						}, false)
						callback(&cloudfront.ListDistributionsOutput{
							DistributionList: &cloudfront.DistributionList{
								Items: []*cloudfront.DistributionSummary{
									{Id: aws.String("distribution4")},
									{Id: aws.String("distribution5")},
									{Id: aws.String("distribution6")},
								},
							},
						}, true)
						return true
					})).Return(nil)
			},
			want: []*cloudfront.DistributionSummary{
				{Id: aws.String("distribution1")},
				{Id: aws.String("distribution2")},
				{Id: aws.String("distribution3")},
				{Id: aws.String("distribution4")},
				{Id: aws.String("distribution5")},
				{Id: aws.String("distribution6")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := awstest.MockFakeCloudFront{}
			tt.mocks(&client)
			r := &cloudfrontRepository{
				client: &client,
			}
			got, err := r.ListAllDistributions()
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
