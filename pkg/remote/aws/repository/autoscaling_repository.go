package repository

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type AutoScalingRepository interface {
	DescribeGroups([]*string) ([]*autoscaling.Group, error)
	DescribeLoadBalancers(string) ([]*autoscaling.LoadBalancerState, error)
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

func (r *autoscalingRepository) DescribeGroups(names []*string) ([]*autoscaling.Group, error) {
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

func (r *autoscalingRepository) DescribeLoadBalancers(autoScalingGroupName string) ([]*autoscaling.LoadBalancerState, error) {
	cacheKey := fmt.Sprintf("autoscalingDescribeLoadBalancers_%s", autoScalingGroupName)
	if v := r.cache.Get(cacheKey); v != nil {
		return v.([]*autoscaling.LoadBalancerState), nil
	}

	input := &autoscaling.DescribeLoadBalancersInput{
		AutoScalingGroupName: aws.String(autoScalingGroupName),
	}
	loadBalancers, err := r.client.DescribeLoadBalancers(input)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, loadBalancers.LoadBalancers)
	return loadBalancers.LoadBalancers, err
}
