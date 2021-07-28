package github

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/github"
)

type GithubRepositoryEnumerator struct {
	repository GithubRepository
	factory    resource.ResourceFactory
}

func NewGithubRepositoryEnumerator(repo GithubRepository, factory resource.ResourceFactory) *GithubRepositoryEnumerator {
	return &GithubRepositoryEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (g *GithubRepositoryEnumerator) SupportedType() resource.ResourceType {
	return github.GithubRepositoryResourceType
}

func (g *GithubRepositoryEnumerator) Enumerate() ([]resource.Resource, error) {
	ids, err := g.repository.ListRepositories()
	if err != nil {
		return nil, remoteerror.NewResourceScanningError(err, string(g.SupportedType()))
	}

	results := make([]resource.Resource, len(ids))

	for _, id := range ids {
		results = append(
			results,
			g.factory.CreateAbstractResource(
				string(g.SupportedType()),
				id,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
