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
			q.Viewer.Repositories.Nodes = []repository{
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
			q.Viewer.Repositories.Nodes = []repository{
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

	assert.Equal([]repository{
		{
			Name: "repo1",
		},
		{
			Name: "repo2",
		},
		{
			Name: "repo3",
		},
		{
			Name: "repo4",
		},
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
			q.Organization.Repositories.Nodes = []repository{
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
			q.Organization.Repositories.Nodes = []repository{
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

	assert.Equal([]repository{
		{
			Name: "repo1",
		},
		{
			Name: "repo2",
		},
		{
			Name: "repo3",
		},
		{
			Name: "repo4",
		},
	}, repos)
}
