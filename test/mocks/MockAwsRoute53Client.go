package mocks

import (
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
)

type ListHostedZonesPagesOutput []struct {
	LastPage bool
	Response *route53.ListHostedZonesOutput
}

type ListResourceRecordSetsPagesOutput []struct {
	LastPage     bool
	Response     *route53.ListResourceRecordSetsOutput
	HostedZoneId string
}

type MockAWSRoute53Client struct {
	route53iface.Route53API
	zonesPages   ListHostedZonesPagesOutput
	recordsPages ListResourceRecordSetsPagesOutput
	err          error
}

func NewMockAWSRoute53ErrorClient(err error) *MockAWSRoute53Client {
	return &MockAWSRoute53Client{err: err}
}

func NewMockAWSRoute53ZoneClient(zonesPages ListHostedZonesPagesOutput, err error) *MockAWSRoute53Client {
	return &MockAWSRoute53Client{zonesPages: zonesPages, err: err}
}

func NewMockAWSRoute53RecordClient(zonesPages ListHostedZonesPagesOutput, recordsPages ListResourceRecordSetsPagesOutput, err error) *MockAWSRoute53Client {
	return &MockAWSRoute53Client{zonesPages: zonesPages, recordsPages: recordsPages, err: err}
}

func (m *MockAWSRoute53Client) ListHostedZonesPages(_ *route53.ListHostedZonesInput, cb func(*route53.ListHostedZonesOutput, bool) bool) error {
	if m.zonesPages == nil && m.err != nil {
		return m.err
	}
	for _, zonesPage := range m.zonesPages {
		cb(zonesPage.Response, zonesPage.LastPage)
	}
	return nil
}

func (m *MockAWSRoute53Client) ListResourceRecordSetsPages(input *route53.ListResourceRecordSetsInput, cb func(*route53.ListResourceRecordSetsOutput, bool) bool) error {
	if m.recordsPages == nil && m.err != nil {
		return m.err
	}
	for _, recordsPage := range m.recordsPages {
		if *input.HostedZoneId == recordsPage.HostedZoneId {
			if shouldContinue := cb(recordsPage.Response, recordsPage.LastPage); !shouldContinue {
				break
			}
		}
	}
	return nil
}
