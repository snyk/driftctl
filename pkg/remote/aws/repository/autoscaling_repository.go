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

	var results []*autoscaling.LaunchConfiguration
	input := &autoscaling.DescribeLaunchConfigurationsInput{}
	err := r.client.DescribeLaunchConfigurationsPages(input, func(resp *autoscaling.DescribeLaunchConfigurationsOutput, lastPage bool) bool {
		results = append(results, resp.LaunchConfigurations...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, results)
	return results, nil
}
