package repository

import (
	"fmt"
	"github.com/snyk/driftctl/enumeration/remote/cache"

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
	cache  cache.Cache
}

func NewRoute53Repository(session *session.Session, c cache.Cache) *route53Repository {
	return &route53Repository{
		route53.New(session),
		c,
	}
}

func (r *route53Repository) ListAllHealthChecks() ([]*route53.HealthCheck, error) {
	if v := r.cache.Get("route53ListAllHealthChecks"); v != nil {
		return v.([]*route53.HealthCheck), nil
	}

	var tables []*route53.HealthCheck
	input := &route53.ListHealthChecksInput{}
	err := r.client.ListHealthChecksPages(input, func(res *route53.ListHealthChecksOutput, lastPage bool) bool {
		tables = append(tables, res.HealthChecks...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	r.cache.Put("route53ListAllHealthChecks", tables)
	return tables, nil
}

func (r *route53Repository) ListAllZones() ([]*route53.HostedZone, error) {
	cacheKey := "route53ListAllZones"
	v := r.cache.GetAndLock(cacheKey)
	defer r.cache.Unlock(cacheKey)
	if v != nil {
		return v.([]*route53.HostedZone), nil
	}

	var result []*route53.HostedZone
	input := &route53.ListHostedZonesInput{}
	err := r.client.ListHostedZonesPages(input, func(res *route53.ListHostedZonesOutput, lastPage bool) bool {
		result = append(result, res.HostedZones...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, result)
	return result, nil
}

func (r *route53Repository) ListRecordsForZone(zoneId string) ([]*route53.ResourceRecordSet, error) {
	cacheKey := fmt.Sprintf("route53ListRecordsForZone_%s", zoneId)
	if v := r.cache.Get(cacheKey); v != nil {
		return v.([]*route53.ResourceRecordSet), nil
	}

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

	r.cache.Put(cacheKey, results)
	return results, nil
}
