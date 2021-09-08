package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type CloudformationRepository interface {
	ListAllStacks() ([]*cloudformation.Stack, error)
}

type cloudformationRepository struct {
	client cloudformationiface.CloudFormationAPI
	cache  cache.Cache
}

func NewCloudformationRepository(session *session.Session, c cache.Cache) *cloudformationRepository {
	return &cloudformationRepository{
		cloudformation.New(session),
		c,
	}
}

func (r *cloudformationRepository) ListAllStacks() ([]*cloudformation.Stack, error) {
	if v := r.cache.Get("cloudformationListAllStacks"); v != nil {
		return v.([]*cloudformation.Stack), nil
	}

	var stacks []*cloudformation.Stack
	input := cloudformation.DescribeStacksInput{}
	err := r.client.DescribeStacksPages(&input,
		func(resp *cloudformation.DescribeStacksOutput, lastPage bool) bool {
			if resp.Stacks != nil {
				stacks = append(stacks, resp.Stacks...)
			}
			return !lastPage
		},
	)
	if err != nil {
		return nil, err
	}

	r.cache.Put("cloudformationListAllStacks", stacks)
	return stacks, nil
}
