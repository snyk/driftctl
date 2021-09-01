package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/cloudfront/cloudfrontiface"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type CloudfrontRepository interface {
	ListAllDistributions() ([]*cloudfront.DistributionSummary, error)
}

type cloudfrontRepository struct {
	*methodLocker
	client cloudfrontiface.CloudFrontAPI
	cache  cache.Cache
}

func NewCloudfrontRepository(session *session.Session, c cache.Cache) *cloudfrontRepository {
	return &cloudfrontRepository{
		newMethodLocker(),
		cloudfront.New(session),
		c,
	}
}

func (r *cloudfrontRepository) ListAllDistributions() ([]*cloudfront.DistributionSummary, error) {
	cacheKey := "cloudfrontListAllDistributions"
	r.Lock(cacheKey)
	defer r.Unlock(cacheKey)

	if v := r.cache.Get(cacheKey); v != nil {
		return v.([]*cloudfront.DistributionSummary), nil
	}

	var distributions []*cloudfront.DistributionSummary
	input := cloudfront.ListDistributionsInput{}
	err := r.client.ListDistributionsPages(&input,
		func(resp *cloudfront.ListDistributionsOutput, lastPage bool) bool {
			if resp.DistributionList != nil {
				distributions = append(distributions, resp.DistributionList.Items...)
			}
			return !lastPage
		},
	)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, distributions)
	return distributions, nil
}
