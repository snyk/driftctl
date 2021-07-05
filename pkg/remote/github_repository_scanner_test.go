package remote

import (
	"testing"

	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	"github.com/cloudskiff/driftctl/pkg/remote/github"
	tf "github.com/cloudskiff/driftctl/pkg/remote/terraform"
	githubres "github.com/cloudskiff/driftctl/pkg/resource/github"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	tftest "github.com/cloudskiff/driftctl/test/terraform"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	"github.com/stretchr/testify/assert"
)

func TestScanGithubRepository(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(client *github.MockGithubRepository)
		err     error
	}{
		{
			test:    "no github repos",
			dirName: "github_repository_empty",
			mocks: func(client *github.MockGithubRepository) {
				client.On("ListRepositories").Return([]string{}, nil)
			},
			err: nil,
		},
		{
			test:    "Multiple github repos Table",
			dirName: "github_repository_multiple",
			mocks: func(client *github.MockGithubRepository) {
				client.On("ListRepositories").Return([]string{
					"driftctl",
					"driftctl-demos",
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

	for _, c := range tests {
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

			remoteLibrary.AddEnumerator(github.NewGithubRepositoryEnumerator(repo, factory, tf.TerraformProviderConfig{
				Name: "test",
			}))
			remoteLibrary.AddDetailsFetcher(githubres.GithubRepositoryResourceType, common.NewGenericDetailsFetcher(githubres.GithubRepositoryResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, githubres.GithubRepositoryResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}
