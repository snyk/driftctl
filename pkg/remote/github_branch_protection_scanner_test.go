package remote

import (
	"testing"

	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	"github.com/cloudskiff/driftctl/pkg/remote/github"
	githubres "github.com/cloudskiff/driftctl/pkg/resource/github"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	tftest "github.com/cloudskiff/driftctl/test/terraform"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	"github.com/stretchr/testify/assert"
)

func TestScanGithubBranchProtection(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(client *github.MockGithubRepository)
		err     error
	}{
		{
			test:    "no branch protection",
			dirName: "github_branch_protection_empty",
			mocks: func(client *github.MockGithubRepository) {
				client.On("ListBranchProtection").Return([]string{}, nil)
			},
			err: nil,
		},
		{
			test:    "Multiple branch protections",
			dirName: "github_branch_protection_multiples",
			mocks: func(client *github.MockGithubRepository) {
				client.On("ListBranchProtection").Return([]string{
					"MDIwOkJyYW5jaFByb3RlY3Rpb25SdWxlMTk1NDg0NzI=", //"repo0:main"
					"MDIwOkJyYW5jaFByb3RlY3Rpb25SdWxlMTk1NDg0Nzg=", //"repo0:toto"
					"MDIwOkJyYW5jaFByb3RlY3Rpb25SdWxlMTk1NDg0NzQ=", //"repo1:main"
					"MDIwOkJyYW5jaFByb3RlY3Rpb25SdWxlMTk1NDg0ODA=", //"repo1:toto"
					"MDIwOkJyYW5jaFByb3RlY3Rpb25SdWxlMTk1NDg0NzE=", //"repo2:main"
					"MDIwOkJyYW5jaFByb3RlY3Rpb25SdWxlMTk1NDg0Nzc=", //"repo2:toto"
				}, nil)
			},
			err: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("github", "4.4.0")
	githubres.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)
	alerter := &mocks.AlerterInterface{}

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			scanOptions := ScannerOptions{Deep: true}

			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			mockedRepo := github.MockGithubRepository{}
			c.mocks(&mockedRepo)
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
			remoteLibrary.AddDetailsFetcher(githubres.GithubBranchProtectionResourceType, common.NewGenericDetailsFetcher(githubres.GithubBranchProtectionResourceType, provider, deserializer))

			s := NewScanner(remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, githubres.GithubBranchProtectionResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}
