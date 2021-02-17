package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
)

type Route53Repository interface {
	ListAllHealthChecks() ([]*route53.HealthCheck, error)
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
