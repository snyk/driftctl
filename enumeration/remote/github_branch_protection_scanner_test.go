package remote

import (
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/remote/github"
	"github.com/snyk/driftctl/enumeration/terraform"

	"github.com/pkg/errors"
	githubres "github.com/snyk/driftctl/enumeration/resource/github"
	"github.com/snyk/driftctl/mocks"

	tftest "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/mock"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/test/goldenfile"
	"github.com/stretchr/testify/assert"
)

func TestScanGithubBranchProtection(t *testing.T) {
	cases := []struct {
		test           string
		dirName        string
		mocks          func(*github.MockGithubRepository, *mocks.AlerterInterface)
		assertExpected func(*testing.T, []*resource.Resource)
		err            error
	}{
		{
			test:    "no branch protection",
			dirName: "github_branch_protection_empty",
			mocks: func(client *github.MockGithubRepository, alerter *mocks.AlerterInterface) {
				client.On("ListBranchProtection").Return([]string{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			err: nil,
		},
		{
			test:    "Multiple branch protections",
			dirName: "github_branch_protection_multiples",
			mocks: func(client *github.MockGithubRepository, alerter *mocks.AlerterInterface) {
				client.On("ListBranchProtection").Return([]string{
					"MDIwOkJyYW5jaFByb3RlY3Rpb25SdWxlMTk1NDg0NzI=", // "repo0:main"
					"MDIwOkJyYW5jaFByb3RlY3Rpb25SdWxlMTk1NDg0Nzg=", // "repo0:toto"
					"MDIwOkJyYW5jaFByb3RlY3Rpb25SdWxlMTk1NDg0NzQ=", // "repo1:main"
					"MDIwOkJyYW5jaFByb3RlY3Rpb25SdWxlMTk1NDg0ODA=", // "repo1:toto"
					"MDIwOkJyYW5jaFByb3RlY3Rpb25SdWxlMTk1NDg0NzE=", // "repo2:main"
					"MDIwOkJyYW5jaFByb3RlY3Rpb25SdWxlMTk1NDg0Nzc=", // "repo2:toto"
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 6)

				assert.Equal(t, "MDIwOkJyYW5jaFByb3RlY3Rpb25SdWxlMTk1NDg0NzI=", got[0].ResourceId())
				assert.Equal(t, githubres.GithubBranchProtectionResourceType, got[0].ResourceType())

				assert.Equal(t, "MDIwOkJyYW5jaFByb3RlY3Rpb25SdWxlMTk1NDg0Nzc=", got[5].ResourceId())
				assert.Equal(t, githubres.GithubBranchProtectionResourceType, got[5].ResourceType())
			},
			err: nil,
		},
		{
			test:    "cannot list branch protections",
			dirName: "github_branch_protection_empty",
			mocks: func(client *github.MockGithubRepository, alerter *mocks.AlerterInterface) {
				client.On("ListBranchProtection").Return(nil, errors.New("Your token has not been granted the required scopes to execute this query."))

				alerter.On("SendAlert", githubres.GithubBranchProtectionResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteGithubTerraform, remoteerr.NewResourceListingErrorWithType(errors.New("Your token has not been granted the required scopes to execute this query."), githubres.GithubBranchProtectionResourceType, githubres.GithubBranchProtectionResourceType), alerts.EnumerationPhase)).Return()
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

			remoteLibrary.AddEnumerator(github.NewGithubBranchProtectionEnumerator(repo, factory))

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
