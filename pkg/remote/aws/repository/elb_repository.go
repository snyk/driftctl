package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/snyk/driftctl/pkg/remote/cache"
)

type ELBRepository interface {
	ListAllLoadBalancers() ([]*elb.LoadBalancerDescription, error)
}

type elbRepository struct {
	client elbiface.ELBAPI
	cache  cache.Cache
}

func NewELBRepository(session *session.Session, c cache.Cache) *elbRepository {
	return &elbRepository{
		elb.New(session),
		c,
	}
}

func (r *elbRepository) ListAllLoadBalancers() ([]*elb.LoadBalancerDescription, error) {
	if v := r.cache.Get("elbListAllLoadBalancers"); v != nil {
		return v.([]*elb.LoadBalancerDescription), nil
	}

	results := make([]*elb.LoadBalancerDescription, 0)
	input := elb.DescribeLoadBalancersInput{}
	err := r.client.DescribeLoadBalancersPages(&input, func(res *elb.DescribeLoadBalancersOutput, lastPage bool) bool {
		results = append(results, res.LoadBalancerDescriptions...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	r.cache.Put("elbListAllLoadBalancers", results)
	return results, nil
}
