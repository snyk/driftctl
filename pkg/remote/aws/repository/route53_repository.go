package repository

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
)

type Route53Repository interface {
	ListAllHealthChecks() ([]*route53.HealthCheck, error)
	ListAllZones() ([]*route53.HostedZone, error)
	ListRecordsForZone(zoneId string) ([]*route53.ResourceRecordSet, error)
}

type route53Repository struct {
	client route53iface.Route53API
}

func NewRoute53Repository(session *session.Session) *route53Repository {
	return &route53Repository{
		route53.New(session),
	}
}

func (r *route53Repository) ListAllHealthChecks() ([]*route53.HealthCheck, error) {
	var tables []*route53.HealthCheck
	input := &route53.ListHealthChecksInput{}
	err := r.client.ListHealthChecksPages(input, func(res *route53.ListHealthChecksOutput, lastPage bool) bool {
		tables = append(tables, res.HealthChecks...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return tables, nil
}

func (r *route53Repository) ListAllZones() ([]*route53.HostedZone, error) {
	var result []*route53.HostedZone
	input := &route53.ListHostedZonesInput{}
	err := r.client.ListHostedZonesPages(input, func(res *route53.ListHostedZonesOutput, lastPage bool) bool {
		result = append(result, res.HostedZones...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *route53Repository) ListRecordsForZone(zoneId string) ([]*route53.ResourceRecordSet, error) {
	var results []*route53.ResourceRecordSet
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneId),
	}
	err := r.client.ListResourceRecordSetsPages(input, func(res *route53.ListResourceRecordSetsOutput, lastPage bool) bool {
		results = append(results, res.ResourceRecordSets...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}
