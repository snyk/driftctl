package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/aws/aws-sdk-go/service/cloudtrail/cloudtrailiface"
	"github.com/snyk/driftctl/enumeration/remote/cache"
)

type CloudtrailRepository interface {
	ListAllTrails() ([]*cloudtrail.TrailInfo, error)
}

type cloudtrailRepository struct {
	client cloudtrailiface.CloudTrailAPI
	cache  cache.Cache
}

func NewCloudtrailRepository(session *session.Session, c cache.Cache) *cloudtrailRepository {
	return &cloudtrailRepository{
		cloudtrail.New(session),
		c,
	}
}

func (r *cloudtrailRepository) ListAllTrails() ([]*cloudtrail.TrailInfo, error) {
	cacheKey := "ListAllTrails"
	if v := r.cache.Get(cacheKey); v != nil {
		return v.([]*cloudtrail.TrailInfo), nil
	}

	var trails []*cloudtrail.TrailInfo
	input := cloudtrail.ListTrailsInput{}
	err := r.client.ListTrailsPages(&input,
		func(resp *cloudtrail.ListTrailsOutput, lastPage bool) bool {
			if resp.Trails != nil {
				trails = append(trails, resp.Trails...)
			}
			return !lastPage
		},
	)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, trails)
	return trails, nil
}
