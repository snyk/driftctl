package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type AutoScalingRepository interface {
	DescribeLaunchConfigurations() ([]*autoscaling.LaunchConfiguration, error)
}

type autoScalingRepository struct {
	client autoscalingiface.AutoScalingAPI
	cache  cache.Cache
}

func NewAutoScalingRepository(session *session.Session, c cache.Cache) *autoScalingRepository {
	return &autoScalingRepository{
		autoscaling.New(session),
		c,
	}
}

func (r *autoScalingRepository) DescribeLaunchConfigurations() ([]*autoscaling.LaunchConfiguration, error) {
	cacheKey := "DescribeLaunchConfigurations"
	if v := r.cache.Get(cacheKey); v != nil {
		return v.([]*autoscaling.LaunchConfiguration), nil
	}

	input := &autoscaling.DescribeLaunchConfigurationsInput{}
	result, err := r.client.DescribeLaunchConfigurations(input)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, result.LaunchConfigurations)
	return result.LaunchConfigurations, nil
}
