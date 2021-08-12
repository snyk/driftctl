package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type ELBRepository interface {
	ListLoadBalancers() ([]*elb.LoadBalancerDescription, error)
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

func (r *elbRepository) ListLoadBalancers() ([]*elb.LoadBalancerDescription, error) {
	if v := r.cache.Get("elbListLoadBalancers"); v != nil {
		return v.([]*elb.LoadBalancerDescription), nil
	}

	input := &elb.DescribeLoadBalancersInput{}
	loadBalancers, err := r.client.DescribeLoadBalancers(input)
	if err != nil {
		return nil, err
	}

	r.cache.Put("elbListLoadBalancers", loadBalancers.LoadBalancerDescriptions)
	return loadBalancers.LoadBalancerDescriptions, err
}
