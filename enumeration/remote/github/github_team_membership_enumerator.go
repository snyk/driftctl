package github

import (
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/github"
)

type GithubTeamMembershipEnumerator struct {
	repository GithubRepository
	factory    resource.ResourceFactory
}

func NewGithubTeamMembershipEnumerator(repo GithubRepository, factory resource.ResourceFactory) *GithubTeamMembershipEnumerator {
	return &GithubTeamMembershipEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (g *GithubTeamMembershipEnumerator) SupportedType() resource.ResourceType {
	return github.GithubTeamMembershipResourceType
}

func (g *GithubTeamMembershipEnumerator) Enumerate() ([]*resource.Resource, error) {
	ids, err := g.repository.ListTeamMemberships()
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
