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

func Test_route53Repository_ListAllZones(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *mocks.Route53Client)
		want    []*route53.HostedZone
		wantErr error
	}{
		{name: "Zones with 2 pages",
			mocks: func(client *mocks.Route53Client) {
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
					})).Return(nil)
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
			client := &mocks.Route53Client{}
			tt.mocks(client)
			r := &route53Repository{
				client: client,
			}
			got, err := r.ListAllZones()
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

func Test_route53Repository_ListRecordsForZone(t *testing.T) {
	tests := []struct {
		name    string
		zoneIds []string
		mocks   func(client *mocks.Route53Client)
		want    []*route53.ResourceRecordSet
		wantErr error
	}{
		{
			name: "records for zone with 2 pages",
			zoneIds: []string{
				"1",
			},
			mocks: func(client *mocks.Route53Client) {
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
					})).Return(nil)
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
			client := &mocks.Route53Client{}
			tt.mocks(client)
			r := &route53Repository{
				client: client,
			}
			for _, id := range tt.zoneIds {
				got, err := r.ListRecordsForZone(id)
				assert.Equal(t, tt.wantErr, err)
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
