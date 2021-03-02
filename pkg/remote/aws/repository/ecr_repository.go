package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
)

type ECRRepository interface {
	ListAllRepositories() ([]*ecr.Repository, error)
}

type ecrRepository struct {
	client ecriface.ECRAPI
}

func NewECRRepository(session *session.Session) *ecrRepository {
	return &ecrRepository{
		ecr.New(session),
	}
}

func (r *ecrRepository) ListAllRepositories() ([]*ecr.Repository, error) {
	var repositories []*ecr.Repository
	input := &ecr.DescribeRepositoriesInput{}
	err := r.client.DescribeRepositoriesPages(input, func(res *ecr.DescribeRepositoriesOutput, lastPage bool) bool {
		repositories = append(repositories, res.Repositories...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return repositories, nil
}
