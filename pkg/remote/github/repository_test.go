package github

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/mocks"
	"github.com/pkg/errors"
	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListRepositoriesForUser_WithError(t *testing.T) {
	assert := assert.New(t)

	mockedClient := mocks.GithubGraphQLClient{}
	expectedError := errors.New("test error from graphql")
	mockedClient.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(expectedError)

	r := githubRepository{
		client: &mockedClient,
		config: githubConfig{},
	}

	_, err := r.ListRepositories()
	assert.Equal(expectedError, err)
}

func TestListRepositoriesForUser(t *testing.T) {

	assert := assert.New(t)

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
		}).Return(nil)

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
		}).Return(nil)

	r := githubRepository{
		client: &mockedClient,
		ctx:    context.TODO(),
		config: githubConfig{},
	}

	repos, err := r.ListRepositories()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal([]string{
		"repo1",
		"repo2",
		"repo3",
		"repo4",
	}, repos)
}

func TestListRepositoriesForOrganization_WithError(t *testing.T) {
	assert := assert.New(t)

	mockedClient := mocks.GithubGraphQLClient{}
	expectedError := errors.New("test error from graphql")
	mockedClient.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(expectedError)

	r := githubRepository{
		client: &mockedClient,
		config: githubConfig{
			Organization: "testorg",
		},
	}

	_, err := r.ListRepositories()
	assert.Equal(expectedError, err)
}

func TestListRepositoriesForOrganization(t *testing.T) {
	assert := assert.New(t)

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
		}).Return(nil)

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
		}).Return(nil)

	r := githubRepository{
		client: &mockedClient,
		ctx:    context.TODO(),
		config: githubConfig{
			Organization: "testorg",
		},
	}

	repos, err := r.ListRepositories()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal([]string{
		"repo1",
		"repo2",
		"repo3",
		"repo4",
	}, repos)
}

func TestListTeams_WithError(t *testing.T) {
	assert := assert.New(t)

	mockedClient := mocks.GithubGraphQLClient{}
	expectedError := errors.New("test error from graphql")
	mockedClient.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(expectedError)

	r := githubRepository{
		client: &mockedClient,
		config: githubConfig{
			Organization: "testorg",
		},
	}

	_, err := r.ListTeams()
	assert.Equal(expectedError, err)
}

func TestListTeams_WithoutOrganization(t *testing.T) {
	assert := assert.New(t)

	r := githubRepository{}

	teams, err := r.ListTeams()
	assert.Nil(err)
	assert.Equal([]Team{}, teams)
}

func TestListTeams(t *testing.T) {
	assert := assert.New(t)

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
		}).Return(nil)

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
		}).Return(nil)

	r := githubRepository{
		client: &mockedClient,
		ctx:    context.TODO(),
		config: githubConfig{
			Organization: "testorg",
		},
	}

	teams, err := r.ListTeams()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal([]Team{
		{1, "1"},
		{2, "2"},
		{3, "3"},
		{4, "4"},
	}, teams)
}

func TestListTeamMemberships_WithTeamListingError(t *testing.T) {
	assert := assert.New(t)

	mockedClient := mocks.GithubGraphQLClient{}
	expectedError := errors.New("test error from graphql")
	mockedClient.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(expectedError)

	r := githubRepository{
		client: &mockedClient,
		config: githubConfig{
			Organization: "testorg",
		},
	}

	_, err := r.ListTeamMemberships()
	assert.Equal(expectedError, err)
}

func TestListTeamMemberships_WithError(t *testing.T) {
	assert := assert.New(t)

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
	}

	_, err := r.ListTeamMemberships()
	assert.Equal(expectedError, err)
}

func TestListTeamMemberships_WithoutOrganization(t *testing.T) {
	assert := assert.New(t)

	r := githubRepository{}

	teams, err := r.ListTeamMemberships()
	assert.Nil(err)
	assert.Equal([]string{}, teams)
}

func TestListTeamMemberships(t *testing.T) {
	assert := assert.New(t)

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
		}).Return(nil)

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
		}).Return(nil)

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
		}).Return(nil)

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
		}).Return(nil)

	r := githubRepository{
		client: &mockedClient,
		ctx:    context.TODO(),
		config: githubConfig{
			Organization: "testorg",
		},
	}

	memberships, err := r.ListTeamMemberships()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal([]string{
		"1:user-1",
		"1:user-2",
		"1:user-3",
		"1:user-4",
		"2:user-5",
		"2:user-6",
	}, memberships)
}

func TestListMembership_WithError(t *testing.T) {
	assert := assert.New(t)

	mockedClient := mocks.GithubGraphQLClient{}
	expectedError := errors.New("test error from graphql")
	mockedClient.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(expectedError)

	r := githubRepository{
		client: &mockedClient,
		config: githubConfig{
			Organization: "testorg",
		},
	}

	_, err := r.ListMembership()
	assert.Equal(expectedError, err)
}

func TestListMembership_WithoutOrganization(t *testing.T) {
	assert := assert.New(t)

	r := githubRepository{}

	teams, err := r.ListMembership()
	assert.Nil(err)
	assert.Equal([]string{}, teams)
}

func TestListMembership(t *testing.T) {
	assert := assert.New(t)

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
		}).Return(nil)

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
		}).Return(nil)

	r := githubRepository{
		client: &mockedClient,
		ctx:    context.TODO(),
		config: githubConfig{
			Organization: "testorg",
		},
	}

	teams, err := r.ListMembership()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal([]string{
		"testorg:user-admin",
		"testorg:user-non-admin-1",
		"testorg:user-non-admin-2",
		"testorg:user-non-admin-3",
	}, teams)
}
