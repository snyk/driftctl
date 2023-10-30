package remote

import (
	"errors"
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/remote/github"
	"github.com/snyk/driftctl/enumeration/terraform"

	githubres "github.com/snyk/driftctl/enumeration/resource/github"
	"github.com/snyk/driftctl/mocks"

	tftest "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/mock"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/test/goldenfile"
	"github.com/stretchr/testify/assert"
)

func TestScanGithubMembership(t *testing.T) {
	cases := []struct {
		test           string
		dirName        string
		mocks          func(*github.MockGithubRepository, *mocks.AlerterInterface)
		assertExpected func(*testing.T, []*resource.Resource)
		err            error
	}{
		{
			test:    "no members",
			dirName: "github_membership_empty",
			mocks: func(client *github.MockGithubRepository, alerter *mocks.AlerterInterface) {
				client.On("ListMembership").Return([]string{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			err: nil,
		},
		{
			test:    "Multiple membership with admin and member roles",
			dirName: "github_membership_multiple",
			mocks: func(client *github.MockGithubRepository, alerter *mocks.AlerterInterface) {
				client.On("ListMembership").Return([]string{
					"driftctl-test:driftctl-acceptance-tester",
					"driftctl-test:eliecharra",
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, "driftctl-test:driftctl-acceptance-tester", got[0].ResourceId())
				assert.Equal(t, githubres.GithubMembershipResourceType, got[0].ResourceType())

				assert.Equal(t, "driftctl-test:eliecharra", got[1].ResourceId())
				assert.Equal(t, githubres.GithubMembershipResourceType, got[1].ResourceType())
			},
			err: nil,
		},
		{
			test:    "cannot list membership",
			dirName: "github_membership_empty",
			mocks: func(client *github.MockGithubRepository, alerter *mocks.AlerterInterface) {
				client.On("ListMembership").Return(nil, errors.New("Your token has not been granted the required scopes to execute this query."))

				alerter.On("SendAlert", githubres.GithubMembershipResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteGithubTerraform, remoteerr.NewResourceListingErrorWithType(errors.New("Your token has not been granted the required scopes to execute this query."), githubres.GithubMembershipResourceType, githubres.GithubMembershipResourceType), alerts.EnumerationPhase)).Return()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			err: nil,
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			mockedRepo := github.MockGithubRepository{}
			c.mocks(&mockedRepo, alerter)

			var repo github.GithubRepository = &mockedRepo

			realProvider, err := tftest.InitTestGithubProvider(providerLibrary, "4.4.0")
			if err != nil {
				t.Fatal(err)
			}
			provider := tftest.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				repo = github.NewGithubRepository(realProvider.GetConfig(), cache.New(0))
			}

			remoteLibrary.AddEnumerator(github.NewGithubMembershipEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.err)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			mockedRepo.AssertExpectations(tt)
			alerter.AssertExpectations(tt)
		})
	}
}
