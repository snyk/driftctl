package github

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type GithubRepository interface {
	ListRepositories() ([]string, error)
	ListTeams() ([]int, error)
	ListMembership() ([]string, error)
}

type GithubGraphQLClient interface {
	Query(ctx context.Context, q interface{}, variables map[string]interface{}) error
}

type githubRepository struct {
	client GithubGraphQLClient
	ctx    context.Context
	config githubConfig
}

func NewGithubRepository(config githubConfig) *githubRepository {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	oauthClient := oauth2.NewClient(ctx, ts)

	repo := &githubRepository{
		client: githubv4.NewClient(oauthClient),
		ctx:    context.Background(),
		config: config,
	}

	return repo
}

func (r *githubRepository) ListRepositories() ([]string, error) {
	if r.config.Organization != "" {
		return r.listRepoForOrg()
	}
	return r.listRepoForOwner()
}

type pageInfo struct {
	EndCursor   githubv4.String
	HasNextPage bool
}

type listRepoForOrgQuery struct {
	Organization struct {
		Repositories struct {
			Nodes []struct {
				Name string
			}
			PageInfo pageInfo
		} `graphql:"repositories(first: 100, after: $cursor)"`
	} `graphql:"organization(login: $org)"`
}

func (r *githubRepository) listRepoForOrg() ([]string, error) {
	query := listRepoForOrgQuery{}
	variables := map[string]interface{}{
		"org":    (githubv4.String)(r.config.Organization),
		"cursor": (*githubv4.String)(nil),
	}
	var results []string
	for {
		err := r.client.Query(r.ctx, &query, variables)
		if err != nil {
			return nil, err
		}
		for _, repo := range query.Organization.Repositories.Nodes {
			results = append(results, repo.Name)
		}
		if !query.Organization.Repositories.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Organization.Repositories.PageInfo.EndCursor)
	}
	return results, nil
}

type listRepoForOwnerQuery struct {
	Viewer struct {
		Repositories struct {
			Nodes []struct {
				Name string
			}
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage bool
			}
		} `graphql:"repositories(first: 100, after: $cursor)"`
	}
}

func (r githubRepository) listRepoForOwner() ([]string, error) {
	query := listRepoForOwnerQuery{}
	variables := map[string]interface{}{
		"cursor": (*githubv4.String)(nil),
	}
	var results []string
	for {
		err := r.client.Query(r.ctx, &query, variables)
		if err != nil {
			return nil, err
		}
		for _, repo := range query.Viewer.Repositories.Nodes {
			results = append(results, repo.Name)
		}
		if !query.Viewer.Repositories.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Viewer.Repositories.PageInfo.EndCursor)
	}
	return results, nil
}

type listTeamsQuery struct {
	Organization struct {
		Teams struct {
			Nodes []struct {
				DatabaseId int
			}
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage bool
			}
		} `graphql:"teams(first: 100, after: $cursor)"`
	} `graphql:"organization(login: $login)"`
}

func (r githubRepository) ListTeams() ([]int, error) {
	query := listTeamsQuery{}
	results := make([]int, 0)
	if r.config.Organization == "" {
		return results, nil
	}
	variables := map[string]interface{}{
		"cursor": (*githubv4.String)(nil),
		"login":  (githubv4.String)(r.config.Organization),
	}
	for {
		err := r.client.Query(r.ctx, &query, variables)
		if err != nil {
			return nil, err
		}
		for _, team := range query.Organization.Teams.Nodes {
			results = append(results, team.DatabaseId)
		}
		if !query.Organization.Teams.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Organization.Teams.PageInfo.EndCursor)
	}
	return results, nil
}

type listMembership struct {
	Organization struct {
		MembersWithRole struct {
			Nodes []struct {
				Login string
			}
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage bool
			}
		} `graphql:"membersWithRole(first: 100, after: $cursor)"`
	} `graphql:"organization(login: $login)"`
}

func (r *githubRepository) ListMembership() ([]string, error) {
	query := listMembership{}
	results := make([]string, 0)
	if r.config.Organization == "" {
		return results, nil
	}
	variables := map[string]interface{}{
		"cursor": (*githubv4.String)(nil),
		"login":  (githubv4.String)(r.config.Organization),
	}
	for {
		err := r.client.Query(r.ctx, &query, variables)
		if err != nil {
			return nil, err
		}
		for _, membership := range query.Organization.MembersWithRole.Nodes {
			results = append(results, fmt.Sprintf("%s:%s", r.config.Organization, membership.Login))
		}
		if !query.Organization.MembersWithRole.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Organization.MembersWithRole.PageInfo.EndCursor)
	}
	return results, nil
}
