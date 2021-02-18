package repository

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/mocks"
	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
)

func Test_route53Repository_ListAllHealthChecks(t *testing.T) {

	tests := []struct {
		name    string
		mocks   func(client *mocks.Route53Client)
		want    []*route53.HealthCheck
		wantErr error
	}{
		{
			name: "List with 2 pages",
			mocks: func(client *mocks.Route53Client) {
				client.On("ListHealthChecksPages",
					&route53.ListHealthChecksInput{},
					mock.MatchedBy(func(callback func(res *route53.ListHealthChecksOutput, lastPage bool) bool) bool {
						callback(&route53.ListHealthChecksOutput{
							HealthChecks: []*route53.HealthCheck{
								{Id: aws.String("1")},
								{Id: aws.String("2")},
								{Id: aws.String("3")},
							},
						}, false)
						callback(&route53.ListHealthChecksOutput{
							HealthChecks: []*route53.HealthCheck{
								{Id: aws.String("4")},
								{Id: aws.String("5")},
								{Id: aws.String("6")},
							},
						}, true)
						return true
					})).Return(nil)
			},
			want: []*route53.HealthCheck{
				{Id: aws.String("1")},
				{Id: aws.String("2")},
				{Id: aws.String("3")},
				{Id: aws.String("4")},
				{Id: aws.String("5")},
				{Id: aws.String("6")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mocks.Route53Client{}
			tt.mocks(client)
			r := &route53Repository{
				client: client,
			}
			got, err := r.ListAllHealthChecks()
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
