package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/snyk/driftctl/pkg/remote/cache"
)

type ELBV2Repository interface {
	ListAllLoadBalancers() ([]*elbv2.LoadBalancer, error)
}

type elbv2Repository struct {
	client elbv2iface.ELBV2API
	cache  cache.Cache
}

func NewELBV2Repository(session *session.Session, c cache.Cache) *elbv2Repository {
	return &elbv2Repository{
		elbv2.New(session),
		c,
	}
}

func (r *elbv2Repository) ListAllLoadBalancers() ([]*elbv2.LoadBalancer, error) {
	if v := r.cache.Get("elbListAllLoadBalancers"); v != nil {
		return v.([]*elbv2.LoadBalancer), nil
	}

	results := make([]*elbv2.LoadBalancer, 0)
	input := &elbv2.DescribeLoadBalancersInput{}
	err := r.client.DescribeLoadBalancersPages(input, func(res *elbv2.DescribeLoadBalancersOutput, lastPage bool) bool {
		results = append(results, res.LoadBalancers...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	r.cache.Put("elbListAllLoadBalancers", results)
	return results, err
}
