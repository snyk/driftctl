package github

import (
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/github"
)

type GithubMembershipEnumerator struct {
	Membership GithubRepository
	factory    resource.ResourceFactory
}

func NewGithubMembershipEnumerator(repo GithubRepository, factory resource.ResourceFactory) *GithubMembershipEnumerator {
	return &GithubMembershipEnumerator{
		Membership: repo,
		factory:    factory,
	}
}

func (g *GithubMembershipEnumerator) SupportedType() resource.ResourceType {
	return github.GithubMembershipResourceType
}

func (g *GithubMembershipEnumerator) Enumerate() ([]*resource.Resource, error) {
	ids, err := g.Membership.ListMembership()
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
