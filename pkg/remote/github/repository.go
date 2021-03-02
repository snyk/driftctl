package github

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type GithubRepository interface {
	ListRepositories() ([]string, error)
	ListTeams() ([]Team, error)
	ListMembership() ([]string, error)
	ListTeamMemberships() ([]string, error)
	ListBranchProtection() ([]string, error)
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
				Slug       string
			}
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage bool
			}
		} `graphql:"teams(first: 100, after: $cursor)"`
	} `graphql:"organization(login: $login)"`
}

type Team struct {
	DatabaseId int
	Slug       string
}

func (r githubRepository) ListTeams() ([]Team, error) {
	query := listTeamsQuery{}
	results := make([]Team, 0)
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
			results = append(results, Team{
				DatabaseId: team.DatabaseId,
				Slug:       team.Slug,
			})
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

type listTeamMembershipsQuery struct {
	Organization struct {
		Team struct {
			Members struct {
				Nodes []struct {
					Login string
				}
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
			} `graphql:"members(first: 100, after: $cursor)"`
		} `graphql:"team(slug: $slug)"`
	} `graphql:"organization(login: $login)"`
}

func (r githubRepository) ListTeamMemberships() ([]string, error) {
	teamList, err := r.ListTeams()
	if err != nil {
		return nil, err
	}

	query := listTeamMembershipsQuery{}
	results := make([]string, 0)
	if r.config.Organization == "" {
		return results, nil
	}
	variables := map[string]interface{}{
		"login": (githubv4.String)(r.config.Organization),
	}

	for _, team := range teamList {
		variables["slug"] = (githubv4.String)(team.Slug)
		variables["cursor"] = (*githubv4.String)(nil)
		for {
			err := r.client.Query(r.ctx, &query, variables)
			if err != nil {
				return nil, err
			}
			for _, membership := range query.Organization.Team.Members.Nodes {
				results = append(results, fmt.Sprintf("%d:%s", team.DatabaseId, membership.Login))
			}
			if !query.Organization.Team.Members.PageInfo.HasNextPage {
				break
			}
			variables["cursor"] = query.Organization.Team.Members.PageInfo.EndCursor
		}
	}
	return results, nil
}

type listBranchProtectionQuery struct {
	Repository struct {
		BranchProtectionRules struct {
			Nodes []struct {
				Id string
			}
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage bool
			}
		} `graphql:"branchProtectionRules(first: 1, after: $cursor)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

func (r *githubRepository) ListBranchProtection() ([]string, error) {

	repoList, err := r.ListRepositories()
	if err != nil {
		return nil, err
	}

	results := make([]string, 0)
	query := listBranchProtectionQuery{}
	variables := map[string]interface{}{
		"cursor": (*githubv4.String)(nil),
		"owner":  (githubv4.String)(r.config.getDefaultOwner()),
		"name":   (githubv4.String)(""),
	}

	for _, repo := range repoList {
		variables["name"] = (githubv4.String)(repo)
		variables["cursor"] = (*githubv4.String)(nil)
		for {
			err := r.client.Query(r.ctx, &query, variables)
			if err != nil {
				return nil, err
			}
			for _, protection := range query.Repository.BranchProtectionRules.Nodes {
				results = append(results, protection.Id)
			}

			variables["cursor"] = query.Repository.BranchProtectionRules.PageInfo.EndCursor

			if !query.Repository.BranchProtectionRules.PageInfo.HasNextPage {
				break
			}
		}

	}
	return results, nil
}
