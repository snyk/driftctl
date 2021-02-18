package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/cloudfront/cloudfrontiface"
)

type CloudfrontRepository interface {
	ListAllDistributions() ([]*cloudfront.DistributionSummary, error)
}

type cloudfrontRepository struct {
	client cloudfrontiface.CloudFrontAPI
}

func NewCloudfrontClient(session *session.Session) *cloudfrontRepository {
	return &cloudfrontRepository{
		cloudfront.New(session),
	}
}

func (r *cloudfrontRepository) ListAllDistributions() ([]*cloudfront.DistributionSummary, error) {
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
	return distributions, nil
}
