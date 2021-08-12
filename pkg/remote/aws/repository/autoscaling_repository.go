package repository

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type AutoScalingRepository interface {
	ListGroups([]*string) ([]*autoscaling.Group, error)
	ListLoadBalancers(string2 string) ([]*autoscaling.LoadBalancerState, error)
}

type autoscalingRepository struct {
	client autoscalingiface.AutoScalingAPI
	cache  cache.Cache
}

func NewAutoScalingRepository(session *session.Session, c cache.Cache) *autoscalingRepository {
	return &autoscalingRepository{
		autoscaling.New(session),
		c,
	}
}

func (r *autoscalingRepository) ListGroups(names []*string) ([]*autoscaling.Group, error) {
	cacheKey := fmt.Sprintf("autoscalingDescribeGroups_%+v", names)
	if v := r.cache.Get(cacheKey); v != nil {
		return v.([]*autoscaling.Group), nil
	}

	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: names,
	}
	groups, err := r.client.DescribeAutoScalingGroups(input)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, groups.AutoScalingGroups)
	return groups.AutoScalingGroups, err
}

func (r *autoscalingRepository) ListLoadBalancers(autoScalingGroupName string) ([]*autoscaling.LoadBalancerState, error) {
	input := &autoscaling.DescribeLoadBalancersInput{
		AutoScalingGroupName: &autoScalingGroupName,
	}
	loadBalancers, err := r.client.DescribeLoadBalancers(input)
	if err != nil {
		return nil, err
	}
	return loadBalancers.LoadBalancers, err
}
