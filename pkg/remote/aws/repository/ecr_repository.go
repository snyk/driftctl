package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type ECRRepository interface {
	ListAllRepositories() ([]*ecr.Repository, error)
}

type ecrRepository struct {
	client ecriface.ECRAPI
	cache  cache.Cache
}

func NewECRRepository(session *session.Session, c cache.Cache) *ecrRepository {
	return &ecrRepository{
		ecr.New(session),
		c,
	}
}

func (r *ecrRepository) ListAllRepositories() ([]*ecr.Repository, error) {
	if v := r.cache.Get("ecrListAllRepositories"); v != nil {
		return v.([]*ecr.Repository), nil
	}

	var repositories []*ecr.Repository
	input := &ecr.DescribeRepositoriesInput{}
	err := r.client.DescribeRepositoriesPages(input, func(res *ecr.DescribeRepositoriesOutput, lastPage bool) bool {
		repositories = append(repositories, res.Repositories...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	r.cache.Put("ecrListAllRepositories", repositories)
	return repositories, nil
}
