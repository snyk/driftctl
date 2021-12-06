package github

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/shurcooL/githubv4"
	"github.com/snyk/driftctl/mocks"
	"github.com/snyk/driftctl/pkg/remote/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListRepositoriesForUser_WithError(t *testing.T) {
	mockedClient := mocks.GithubGraphQLClient{}
	expectedError := errors.New("test error from graphql")
	mockedClient.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(expectedError)

	r := githubRepository{
		client: &mockedClient,
		config: githubConfig{},
		cache:  cache.New(1),
	}

	_, err := r.ListRepositories()
	assert.Equal(t, expectedError, err)
}

func TestListRepositoriesForUser(t *testing.T) {
	mockedClient := mocks.GithubGraphQLClient{}
	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listRepoForOwnerQuery)
			if !ok {
				return false
			}
			q.Viewer.Repositories.Nodes = []struct{ Name string }{
				{
					Name: "repo1",
				},
				{
					Name: "repo2",
				},
			}
			q.Viewer.Repositories.PageInfo = pageInfo{
				EndCursor:   "next",
				HasNextPage: true,
			}
			return true
		}),
		map[string]interface{}{
			"cursor": (*githubv4.String)(nil),
		}).Return(nil).Once()

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listRepoForOwnerQuery)
			if !ok {
				return false
			}
			q.Viewer.Repositories.Nodes = []struct{ Name string }{
				{
					Name: "repo3",
				},
				{
					Name: "repo4",
				},
			}
			q.Viewer.Repositories.PageInfo = pageInfo{
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"cursor": githubv4.NewString("next"),
		}).Return(nil).Once()

	store := cache.New(1)
	r := githubRepository{
		client: &mockedClient,
		ctx:    context.TODO(),
		config: githubConfig{},
		cache:  store,
	}

	repos, err := r.ListRepositories()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, []string{
		"repo1",
		"repo2",
		"repo3",
		"repo4",
	}, repos)

	// Check that results were cached
	cachedData, err := r.ListRepositories()
	assert.NoError(t, err)
	assert.Equal(t, repos, cachedData)
	assert.IsType(t, []string{}, store.Get("githubListRepositories"))
}

func TestListRepositoriesForOrganization_WithError(t *testing.T) {
	mockedClient := mocks.GithubGraphQLClient{}
	expectedError := errors.New("test error from graphql")
	mockedClient.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(expectedError)

	r := githubRepository{
		client: &mockedClient,
		config: githubConfig{
			Organization: "testorg",
		},
		cache: cache.New(1),
	}

	_, err := r.ListRepositories()
	assert.Equal(t, expectedError, err)
}

func TestListRepositoriesForOrganization(t *testing.T) {
	mockedClient := mocks.GithubGraphQLClient{}
	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listRepoForOrgQuery)
			if !ok {
				return false
			}
			q.Organization.Repositories.Nodes = []struct {
				Name string
			}{
				{
					Name: "repo1",
				},
				{
					Name: "repo2",
				},
			}
			q.Organization.Repositories.PageInfo = pageInfo{
				EndCursor:   "next",
				HasNextPage: true,
			}
			return true
		}),
		map[string]interface{}{
			"org":    (githubv4.String)("testorg"),
			"cursor": (*githubv4.String)(nil),
		}).Return(nil).Once()

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listRepoForOrgQuery)
			if !ok {
				return false
			}
			q.Organization.Repositories.Nodes = []struct {
				Name string
			}{
				{
					Name: "repo3",
				},
				{
					Name: "repo4",
				},
			}
			q.Organization.Repositories.PageInfo = pageInfo{
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"org":    (githubv4.String)("testorg"),
			"cursor": githubv4.NewString("next"),
		}).Return(nil).Once()

	store := cache.New(1)
	r := githubRepository{
		client: &mockedClient,
		ctx:    context.TODO(),
		config: githubConfig{
			Organization: "testorg",
		},
		cache: store,
	}

	repos, err := r.ListRepositories()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, []string{
		"repo1",
		"repo2",
		"repo3",
		"repo4",
	}, repos)

	// Check that results were cached
	cachedData, err := r.ListRepositories()
	assert.NoError(t, err)
	assert.Equal(t, repos, cachedData)
	assert.IsType(t, []string{}, store.Get("githubListRepositories"))
}

func TestListTeams_WithError(t *testing.T) {
	mockedClient := mocks.GithubGraphQLClient{}
	expectedError := errors.New("test error from graphql")
	mockedClient.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(expectedError)

	r := githubRepository{
		client: &mockedClient,
		config: githubConfig{
			Organization: "testorg",
		},
		cache: cache.New(1),
	}

	_, err := r.ListTeams()
	assert.Equal(t, expectedError, err)
}

func TestListTeams_WithoutOrganization(t *testing.T) {
	r := githubRepository{cache: cache.New(1)}

	teams, err := r.ListTeams()
	assert.Nil(t, err)
	assert.Equal(t, []Team{}, teams)
}

func TestListTeams(t *testing.T) {
	mockedClient := mocks.GithubGraphQLClient{}
	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listTeamsQuery)
			if !ok {
				return false
			}
			q.Organization.Teams.Nodes = []struct {
				DatabaseId int
				Slug       string
			}{
				{
					DatabaseId: 1,
					Slug:       "1",
				},
				{
					DatabaseId: 2,
					Slug:       "2",
				},
			}
			q.Organization.Teams.PageInfo = pageInfo{
				EndCursor:   "next",
				HasNextPage: true,
			}
			return true
		}),
		map[string]interface{}{
			"login":  (githubv4.String)("testorg"),
			"cursor": (*githubv4.String)(nil),
		}).Return(nil).Once()

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listTeamsQuery)
			if !ok {
				return false
			}
			q.Organization.Teams.Nodes = []struct {
				DatabaseId int
				Slug       string
			}{
				{
					DatabaseId: 3,
					Slug:       "3",
				},
				{
					DatabaseId: 4,
					Slug:       "4",
				},
			}
			q.Organization.Teams.PageInfo = pageInfo{
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"login":  (githubv4.String)("testorg"),
			"cursor": githubv4.NewString("next"),
		}).Return(nil).Once()

	store := cache.New(1)
	r := githubRepository{
		client: &mockedClient,
		ctx:    context.TODO(),
		config: githubConfig{
			Organization: "testorg",
		},
		cache: store,
	}

	teams, err := r.ListTeams()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, []Team{
		{1, "1"},
		{2, "2"},
		{3, "3"},
		{4, "4"},
	}, teams)

	// Check that results were cached
	cachedData, err := r.ListTeams()
	assert.NoError(t, err)
	assert.Equal(t, teams, cachedData)
	assert.IsType(t, []Team{}, store.Get("githubListTeams"))
}

func TestListTeamMemberships_WithTeamListingError(t *testing.T) {
	mockedClient := mocks.GithubGraphQLClient{}
	expectedError := errors.New("test error from graphql")
	mockedClient.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(expectedError)

	r := githubRepository{
		client: &mockedClient,
		config: githubConfig{
			Organization: "testorg",
		},
		cache: cache.New(1),
	}

	_, err := r.ListTeamMemberships()
	assert.Equal(t, expectedError, err)
}

func TestListTeamMemberships_WithError(t *testing.T) {
	mockedClient := mocks.GithubGraphQLClient{}

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listTeamsQuery)
			if !ok {
				return false
			}
			q.Organization.Teams.Nodes = []struct {
				DatabaseId int
				Slug       string
			}{
				{
					DatabaseId: 1,
					Slug:       "foo",
				},
			}
			q.Organization.Teams.PageInfo = pageInfo{
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"login":  (githubv4.String)("testorg"),
			"cursor": (*githubv4.String)(nil),
		}).Return(nil)

	expectedError := errors.New("test error from graphql")
	mockedClient.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(expectedError)

	r := githubRepository{
		client: &mockedClient,
		config: githubConfig{
			Organization: "testorg",
		},
		cache: cache.New(1),
	}

	_, err := r.ListTeamMemberships()
	assert.Equal(t, expectedError, err)
}

func TestListTeamMemberships_WithoutOrganization(t *testing.T) {
	r := githubRepository{cache: cache.New(1)}

	teams, err := r.ListTeamMemberships()
	assert.Nil(t, err)
	assert.Equal(t, []string{}, teams)
}

func TestListTeamMemberships(t *testing.T) {
	mockedClient := mocks.GithubGraphQLClient{}
	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listTeamsQuery)
			if !ok {
				return false
			}
			q.Organization.Teams.Nodes = []struct {
				DatabaseId int
				Slug       string
			}{
				{
					DatabaseId: 1,
					Slug:       "foo",
				},
				{
					DatabaseId: 2,
					Slug:       "bar",
				},
			}
			q.Organization.Teams.PageInfo = pageInfo{
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"login":  (githubv4.String)("testorg"),
			"cursor": (*githubv4.String)(nil),
		}).Return(nil).Once()

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listTeamMembershipsQuery)
			if !ok {
				return false
			}
			q.Organization.Team.Members.Nodes = []struct {
				Login string
			}{
				{
					Login: "user-1",
				},
				{
					Login: "user-2",
				},
			}
			q.Organization.Team.Members.PageInfo = pageInfo{
				EndCursor:   "next",
				HasNextPage: true,
			}
			return true
		}),
		map[string]interface{}{
			"login":  (githubv4.String)("testorg"),
			"cursor": (*githubv4.String)(nil),
			"slug":   (githubv4.String)("foo"),
		}).Return(nil).Once()

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listTeamMembershipsQuery)
			if !ok {
				return false
			}
			q.Organization.Team.Members.Nodes = []struct {
				Login string
			}{
				{
					Login: "user-3",
				},
				{
					Login: "user-4",
				},
			}
			q.Organization.Team.Members.PageInfo = pageInfo{
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"login":  (githubv4.String)("testorg"),
			"cursor": (githubv4.String)("next"),
			"slug":   (githubv4.String)("foo"),
		}).Return(nil).Once()

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listTeamMembershipsQuery)
			if !ok {
				return false
			}
			q.Organization.Team.Members.Nodes = []struct {
				Login string
			}{
				{
					Login: "user-5",
				},
				{
					Login: "user-6",
				},
			}
			q.Organization.Team.Members.PageInfo = pageInfo{
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"login":  (githubv4.String)("testorg"),
			"cursor": (*githubv4.String)(nil),
			"slug":   (githubv4.String)("bar"),
		}).Return(nil).Once()

	store := cache.New(1)
	r := githubRepository{
		client: &mockedClient,
		ctx:    context.TODO(),
		config: githubConfig{
			Organization: "testorg",
		},
		cache: store,
	}

	memberships, err := r.ListTeamMemberships()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, []string{
		"1:user-1",
		"1:user-2",
		"1:user-3",
		"1:user-4",
		"2:user-5",
		"2:user-6",
	}, memberships)

	// Check that results were cached
	cachedData, err := r.ListTeamMemberships()
	assert.NoError(t, err)
	assert.Equal(t, memberships, cachedData)
	assert.IsType(t, []string{}, store.Get("githubListTeamMemberships"))
}

func TestListMembership_WithError(t *testing.T) {
	mockedClient := mocks.GithubGraphQLClient{}
	expectedError := errors.New("test error from graphql")
	mockedClient.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(expectedError)

	r := githubRepository{
		client: &mockedClient,
		config: githubConfig{
			Organization: "testorg",
		},
		cache: cache.New(1),
	}

	_, err := r.ListMembership()
	assert.Equal(t, expectedError, err)
}

func TestListMembership_WithoutOrganization(t *testing.T) {
	r := githubRepository{cache: cache.New(1)}

	teams, err := r.ListMembership()
	assert.Nil(t, err)
	assert.Equal(t, []string{}, teams)
}

func TestListMembership(t *testing.T) {
	mockedClient := mocks.GithubGraphQLClient{}
	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listMembership)
			if !ok {
				return false
			}
			q.Organization.MembersWithRole.Nodes = []struct {
				Login string
			}{
				{
					Login: "user-admin",
				},
				{
					Login: "user-non-admin-1",
				},
			}
			q.Organization.MembersWithRole.PageInfo = pageInfo{
				EndCursor:   "next",
				HasNextPage: true,
			}
			return true
		}),
		map[string]interface{}{
			"login":  (githubv4.String)("testorg"),
			"cursor": (*githubv4.String)(nil),
		}).Return(nil).Once()

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listMembership)
			if !ok {
				return false
			}
			q.Organization.MembersWithRole.Nodes = []struct {
				Login string
			}{
				{
					Login: "user-non-admin-2",
				},
				{
					Login: "user-non-admin-3",
				},
			}
			q.Organization.MembersWithRole.PageInfo = pageInfo{
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"login":  (githubv4.String)("testorg"),
			"cursor": githubv4.NewString("next"),
		}).Return(nil).Once()

	store := cache.New(1)
	r := githubRepository{
		client: &mockedClient,
		ctx:    context.TODO(),
		config: githubConfig{
			Organization: "testorg",
		},
		cache: store,
	}

	teams, err := r.ListMembership()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, []string{
		"testorg:user-admin",
		"testorg:user-non-admin-1",
		"testorg:user-non-admin-2",
		"testorg:user-non-admin-3",
	}, teams)

	// Check that results were cached
	cachedData, err := r.ListMembership()
	assert.NoError(t, err)
	assert.Equal(t, teams, cachedData)
	assert.IsType(t, []string{}, store.Get("githubListMembership"))

}

func TestListBranchProtection_WithRepoListingError(t *testing.T) {
	mockedClient := mocks.GithubGraphQLClient{}
	expectedError := errors.New("test error from graphql")
	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listRepoForOrgQuery)
			if !ok {
				return false
			}
			q.Organization.Repositories.Nodes = []struct {
				Name string
			}{
				{
					Name: "repo1",
				},
				{
					Name: "repo2",
				},
			}
			q.Organization.Repositories.PageInfo = pageInfo{
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"org":    (githubv4.String)("my-organization"),
			"cursor": (*githubv4.String)(nil),
		}).Return(expectedError)

	r := githubRepository{
		client: &mockedClient,
		config: githubConfig{
			Organization: "my-organization",
		},
		cache: cache.New(1),
	}

	_, err := r.ListBranchProtection()
	assert.Equal(t, expectedError, err)
}

func TestListBranchProtection_WithError(t *testing.T) {
	mockedClient := mocks.GithubGraphQLClient{}
	expectedError := errors.New("test error from graphql")
	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listRepoForOrgQuery)
			if !ok {
				return false
			}
			q.Organization.Repositories.Nodes = []struct {
				Name string
			}{
				{
					Name: "repo1",
				},
				{
					Name: "repo2",
				},
			}
			q.Organization.Repositories.PageInfo = pageInfo{
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"org":    (githubv4.String)("testorg"),
			"cursor": (*githubv4.String)(nil),
		}).Return(nil)

	mockedClient.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(expectedError)

	r := githubRepository{
		client: &mockedClient,
		config: githubConfig{
			Organization: "testorg",
		},
		cache: cache.New(1),
	}

	_, err := r.ListBranchProtection()
	assert.Equal(t, expectedError, err)
}

func TestListBranchProtection(t *testing.T) {
	mockedClient := mocks.GithubGraphQLClient{}

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listRepoForOrgQuery)
			if !ok {
				return false
			}
			q.Organization.Repositories.Nodes = []struct {
				Name string
			}{
				{
					Name: "repo1",
				},
				{
					Name: "repo2",
				},
			}
			q.Organization.Repositories.PageInfo = pageInfo{
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"org":    (githubv4.String)("my-organization"),
			"cursor": (*githubv4.String)(nil),
		}).Return(nil).Once()

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listBranchProtectionQuery)
			if !ok {
				return false
			}
			q.Repository.BranchProtectionRules.Nodes = []struct {
				Id string
			}{
				{
					Id: "id1",
				},
				{
					Id: "id2",
				},
			}
			q.Repository.BranchProtectionRules.PageInfo = pageInfo{
				EndCursor:   "nextPage",
				HasNextPage: true,
			}
			return true
		}),
		map[string]interface{}{
			"owner":  (githubv4.String)("my-organization"),
			"name":   (githubv4.String)("repo1"),
			"cursor": (*githubv4.String)(nil),
		}).Return(nil).Once()

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listBranchProtectionQuery)
			if !ok {
				return false
			}
			q.Repository.BranchProtectionRules.Nodes = []struct {
				Id string
			}{
				{
					Id: "id3",
				},
				{
					Id: "id4",
				},
			}
			q.Repository.BranchProtectionRules.PageInfo = pageInfo{
				EndCursor:   "nextPage",
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"owner":  (githubv4.String)("my-organization"),
			"name":   (githubv4.String)("repo1"),
			"cursor": (githubv4.String)("nextPage"),
		}).Return(nil).Once()

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listBranchProtectionQuery)
			if !ok {
				return false
			}
			q.Repository.BranchProtectionRules.Nodes = []struct {
				Id string
			}{
				{
					Id: "id5",
				},
				{
					Id: "id6",
				},
			}
			q.Repository.BranchProtectionRules.PageInfo = pageInfo{
				EndCursor:   "nextPage",
				HasNextPage: true,
			}
			return true
		}),
		map[string]interface{}{
			"owner":  (githubv4.String)("my-organization"),
			"name":   (githubv4.String)("repo2"),
			"cursor": (*githubv4.String)(nil),
		}).Return(nil).Once()

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listBranchProtectionQuery)
			if !ok {
				return false
			}
			q.Repository.BranchProtectionRules.Nodes = []struct {
				Id string
			}{
				{
					Id: "id7",
				},
				{
					Id: "id8",
				},
			}
			q.Repository.BranchProtectionRules.PageInfo = pageInfo{
				EndCursor:   "nextPage",
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"owner":  (githubv4.String)("my-organization"),
			"name":   (githubv4.String)("repo2"),
			"cursor": (githubv4.String)("nextPage"),
		}).Return(nil).Once()

	store := cache.New(1)
	r := githubRepository{
		client: &mockedClient,
		ctx:    context.TODO(),
		config: githubConfig{
			Organization: "my-organization",
		},
		cache: store,
	}

	teams, err := r.ListBranchProtection()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, []string{
		"id1",
		"id2",
		"id3",
		"id4",
		"id5",
		"id6",
		"id7",
		"id8",
	}, teams)

	// Check that results were cached
	cachedData, err := r.ListBranchProtection()
	assert.NoError(t, err)
	assert.Equal(t, teams, cachedData)
	assert.IsType(t, []string{}, store.Get("githubListBranchProtection"))
}
