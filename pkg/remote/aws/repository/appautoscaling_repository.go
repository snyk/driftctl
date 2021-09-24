package repository

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling/applicationautoscalingiface"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type AppAutoScalingRepository interface {
	ServiceNamespaceValues() []string
	DescribeScalableTargets(string) ([]*applicationautoscaling.ScalableTarget, error)
	DescribeScalingPolicies(string) ([]*applicationautoscaling.ScalingPolicy, error)
}

type appAutoScalingRepository struct {
	client applicationautoscalingiface.ApplicationAutoScalingAPI
	cache  cache.Cache
}

func NewAppAutoScalingRepository(session *session.Session, c cache.Cache) *appAutoScalingRepository {
	return &appAutoScalingRepository{
		applicationautoscaling.New(session),
		c,
	}
}

func (r *appAutoScalingRepository) ServiceNamespaceValues() []string {
	return applicationautoscaling.ServiceNamespace_Values()
}

func (r *appAutoScalingRepository) DescribeScalableTargets(namespace string) ([]*applicationautoscaling.ScalableTarget, error) {
	cacheKey := fmt.Sprintf("appAutoScalingDescribeScalableTargets_%s", namespace)
	if v := r.cache.Get(cacheKey); v != nil {
		return v.([]*applicationautoscaling.ScalableTarget), nil
	}

	input := &applicationautoscaling.DescribeScalableTargetsInput{
		ServiceNamespace: &namespace,
	}
	result, err := r.client.DescribeScalableTargets(input)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, result.ScalableTargets)
	return result.ScalableTargets, nil
}

func (r *appAutoScalingRepository) DescribeScalingPolicies(namespace string) ([]*applicationautoscaling.ScalingPolicy, error) {
	cacheKey := fmt.Sprintf("appAutoScalingDescribeScalingPolicies_%s", namespace)
	if v := r.cache.Get(cacheKey); v != nil {
		return v.([]*applicationautoscaling.ScalingPolicy), nil
	}

	input := &applicationautoscaling.DescribeScalingPoliciesInput{
		ServiceNamespace: &namespace,
	}
	result, err := r.client.DescribeScalingPolicies(input)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, result.ScalingPolicies)
	return result.ScalingPolicies, nil
}
