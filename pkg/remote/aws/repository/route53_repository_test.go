package repository

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	awstest "github.com/cloudskiff/driftctl/test/aws"
	"github.com/stretchr/testify/mock"

	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
)

func Test_route53Repository_ListAllHealthChecks(t *testing.T) {

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeRoute53)
		want    []*route53.HealthCheck
		wantErr error
	}{
		{
			name: "List with 2 pages",
			mocks: func(client *awstest.MockFakeRoute53) {
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
					})).Return(nil).Once()
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
			store := cache.New(1)
			client := awstest.MockFakeRoute53{}
			tt.mocks(&client)
			r := &route53Repository{
				client: &client,
				cache:  store,
			}
			got, err := r.ListAllHealthChecks()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllHealthChecks()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*route53.HealthCheck{}, store.Get("route53ListAllHealthChecks"))
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

func Test_route53Repository_ListAllZones(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeRoute53)
		want    []*route53.HostedZone
		wantErr error
	}{
		{name: "Zones with 2 pages",
			mocks: func(client *awstest.MockFakeRoute53) {
				client.On("ListHostedZonesPages",
					&route53.ListHostedZonesInput{},
					mock.MatchedBy(func(callback func(res *route53.ListHostedZonesOutput, lastPage bool) bool) bool {
						callback(&route53.ListHostedZonesOutput{
							HostedZones: []*route53.HostedZone{
								{Id: aws.String("1")},
								{Id: aws.String("2")},
								{Id: aws.String("3")},
							},
						}, false)
						callback(&route53.ListHostedZonesOutput{
							HostedZones: []*route53.HostedZone{
								{Id: aws.String("4")},
								{Id: aws.String("5")},
								{Id: aws.String("6")},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			want: []*route53.HostedZone{
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
			store := cache.New(1)
			client := awstest.MockFakeRoute53{}
			tt.mocks(&client)
			r := &route53Repository{
				client: &client,
				cache:  store,
			}
			got, err := r.ListAllZones()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllZones()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*route53.HostedZone{}, store.Get("route53ListAllZones"))
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

func Test_route53Repository_ListRecordsForZone(t *testing.T) {
	tests := []struct {
		name    string
		zoneIds []string
		mocks   func(client *awstest.MockFakeRoute53)
		want    []*route53.ResourceRecordSet
		wantErr error
	}{
		{
			name: "records for zone with 2 pages",
			zoneIds: []string{
				"1",
			},
			mocks: func(client *awstest.MockFakeRoute53) {
				client.On("ListResourceRecordSetsPages",
					&route53.ListResourceRecordSetsInput{
						HostedZoneId: aws.String("1"),
					},
					mock.MatchedBy(func(callback func(res *route53.ListResourceRecordSetsOutput, lastPage bool) bool) bool {
						callback(&route53.ListResourceRecordSetsOutput{
							ResourceRecordSets: []*route53.ResourceRecordSet{
								{Name: aws.String("1")},
								{Name: aws.String("2")},
								{Name: aws.String("3")},
							},
						}, false)
						callback(&route53.ListResourceRecordSetsOutput{
							ResourceRecordSets: []*route53.ResourceRecordSet{
								{Name: aws.String("4")},
								{Name: aws.String("5")},
								{Name: aws.String("6")},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			want: []*route53.ResourceRecordSet{
				{Name: aws.String("1")},
				{Name: aws.String("2")},
				{Name: aws.String("3")},
				{Name: aws.String("4")},
				{Name: aws.String("5")},
				{Name: aws.String("6")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := awstest.MockFakeRoute53{}
			tt.mocks(&client)
			r := &route53Repository{
				client: &client,
				cache:  store,
			}
			for _, id := range tt.zoneIds {
				got, err := r.ListRecordsForZone(id)
				assert.Equal(t, tt.wantErr, err)

				if err == nil {
					// Check that results were cached
					cachedData, err := r.ListRecordsForZone(id)
					assert.NoError(t, err)
					assert.Equal(t, got, cachedData)
					assert.IsType(t, []*route53.ResourceRecordSet{}, store.Get(fmt.Sprintf("route53ListRecordsForZone_%s", id)))
				}

				changelog, err := diff.Diff(got, tt.want)
				assert.Nil(t, err)
				if len(changelog) > 0 {
					for _, change := range changelog {
						t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
					}
					t.Fail()
				}
			}
		})
	}
}
