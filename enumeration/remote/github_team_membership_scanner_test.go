package remote

import (
	"errors"
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	github2 "github.com/snyk/driftctl/enumeration/remote/github"
	terraform2 "github.com/snyk/driftctl/enumeration/terraform"

	githubres "github.com/snyk/driftctl/enumeration/resource/github"
	"github.com/snyk/driftctl/mocks"

	testresource "github.com/snyk/driftctl/test/resource"
	tftest "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/mock"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/goldenfile"
	"github.com/stretchr/testify/assert"
)

func TestScanGithubTeamMembership(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(*github2.MockGithubRepository, *mocks.AlerterInterface)
		err     error
	}{
		{
			test:    "no github team memberships",
			dirName: "github_team_membership_empty",
			mocks: func(client *github2.MockGithubRepository, alerter *mocks.AlerterInterface) {
				client.On("ListTeamMemberships").Return([]string{}, nil)
			},
			err: nil,
		},
		{
			test:    "multiple github team memberships",
			dirName: "github_team_membership_multiple",
			mocks: func(client *github2.MockGithubRepository, alerter *mocks.AlerterInterface) {
				client.On("ListTeamMemberships").Return([]string{
					"4570529:driftctl-acceptance-tester",
					"4570529:wbeuil",
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list team membership",
			dirName: "github_team_membership_empty",
			mocks: func(client *github2.MockGithubRepository, alerter *mocks.AlerterInterface) {
				client.On("ListTeamMemberships").Return(nil, errors.New("Your token has not been granted the required scopes to execute this query."))

				alerter.On("SendAlert", githubres.GithubTeamMembershipResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteGithubTerraform, remoteerr.NewResourceListingErrorWithType(errors.New("Your token has not been granted the required scopes to execute this query."), githubres.GithubTeamMembershipResourceType, githubres.GithubTeamMembershipResourceType), alerts.EnumerationPhase)).Return()
			},
			err: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("github", "4.4.0")
	githubres.InitResourcesMetadata(schemaRepository)
	factory := terraform2.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			scanOptions := ScannerOptions{Deep: true}

			providerLibrary := terraform2.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			mockedRepo := github2.MockGithubRepository{}
			c.mocks(&mockedRepo, alerter)

			var repo github2.GithubRepository = &mockedRepo

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
				repo = github2.NewGithubRepository(realProvider.GetConfig(), cache.New(0))
			}

			remoteLibrary.AddEnumerator(github2.NewGithubTeamMembershipEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(githubres.GithubTeamMembershipResourceType, common.NewGenericDetailsFetcher(githubres.GithubTeamMembershipResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, githubres.GithubTeamMembershipResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			mockedRepo.AssertExpectations(tt)
			alerter.AssertExpectations(tt)
		})
	}
}
