package github

import (
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/github"
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

func (g *GithubBranchProtectionEnumerator) Enumerate() ([]*resource.Resource, error) {
	ids, err := g.repository.ListBranchProtection()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(g.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(ids))

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
