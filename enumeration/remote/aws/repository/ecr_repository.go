package repository

import (
	"fmt"
	"github.com/snyk/driftctl/enumeration/remote/cache"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
)

type ECRRepository interface {
	ListAllRepositories() ([]*ecr.Repository, error)
	GetRepositoryPolicy(*ecr.Repository) (*ecr.GetRepositoryPolicyOutput, error)
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

func (r *ecrRepository) GetRepositoryPolicy(repo *ecr.Repository) (*ecr.GetRepositoryPolicyOutput, error) {
	cacheKey := fmt.Sprintf("ecrListAllRepositoriesGetRepositoryPolicy_%s_%s", *repo.RegistryId, *repo.RepositoryName)
	if v := r.cache.Get(cacheKey); v != nil {
		return v.(*ecr.GetRepositoryPolicyOutput), nil
	}

	var repositoryPolicyInput *ecr.GetRepositoryPolicyInput = &ecr.GetRepositoryPolicyInput{
		RegistryId:     repo.RegistryId,
		RepositoryName: repo.RepositoryName,
	}

	repoOutput, err := r.client.GetRepositoryPolicy(repositoryPolicyInput)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, repoOutput)
	return repoOutput, nil
}
