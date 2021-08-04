package github

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/github"
)

type GithubBranchProtectionEnumerator struct {
	repository GithubRepository
	factory    resource.ResourceFactory
}

func NewGithubBranchProtectionEnumerator(repo GithubRepository, factory resource.ResourceFactory) *GithubBranchProtectionEnumerator {
	return &GithubBranchProtectionEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (g *GithubBranchProtectionEnumerator) SupportedType() resource.ResourceType {
	return github.GithubBranchProtectionResourceType
}

func (g *GithubBranchProtectionEnumerator) Enumerate() ([]resource.Resource, error) {
	ids, err := g.repository.ListBranchProtection()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(g.SupportedType()))
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
