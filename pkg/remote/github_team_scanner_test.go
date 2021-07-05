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

func TestScanGithubTeam(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *github.MockGithubRepository)
		err     error
	}{
		{
			test:    "no github teams",
			dirName: "github_teams_empty",
			mocks: func(client *github.MockGithubRepository) {
				client.On("ListTeams").Return([]github.Team{}, nil)
			},
			err: nil,
		},
		{
			test:    "Multiple github teams with parent",
			dirName: "github_teams_multiple",
			mocks: func(client *github.MockGithubRepository) {
				client.On("ListTeams").Return([]github.Team{
					{DatabaseId: 4556811}, // github_team.team1
					{DatabaseId: 4556812}, // github_team.team2
					{DatabaseId: 4556814}, // github_team.with_parent
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

			remoteLibrary.AddEnumerator(github.NewGithubTeamEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(githubres.GithubTeamResourceType, common.NewGenericDetailsFetcher(githubres.GithubTeamResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, githubres.GithubTeamResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}
