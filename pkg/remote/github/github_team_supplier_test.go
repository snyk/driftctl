package github

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourcegithub "github.com/cloudskiff/driftctl/pkg/resource/github"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	dritftctlmocks "github.com/cloudskiff/driftctl/test/mocks"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGithubTeamSupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *MockGithubRepository)
		err     error
	}{
		{
			test:    "no github teams",
			dirName: "github_teams_empty",
			mocks: func(client *MockGithubRepository) {
				client.On("ListTeams").Return([]Team{}, nil)
			},
			err: nil,
		},
		{
			test:    "Multiple github teams with parent",
			dirName: "github_teams_multiple",
			mocks: func(client *MockGithubRepository) {
				client.On("ListTeams").Return([]Team{
					{DatabaseId: 4556811}, // github_team.team1
					{DatabaseId: 4556812}, // github_team.team2
					{DatabaseId: 4556814}, // github_team.with_parent
				}, nil)
			},
			err: nil,
		},
	}
	for _, c := range cases {
		shouldUpdate := c.dirName == *goldenfile.Update

		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		repo := testresource.InitFakeSchemaRepository(terraform.GITHUB, "4.4.0")
		resourcegithub.InitResourcesMetadata(repo)
		factory := terraform.NewTerraformResourceFactory(repo)

		deserializer := resource.NewDeserializer(factory)

		mockedRepo := MockGithubRepository{}
		c.mocks(&mockedRepo)

		if shouldUpdate {
			provider, err := InitTestGithubProvider(providerLibrary)
			if err != nil {
				t.Fatal(err)
			}

			supplierLibrary.AddSupplier(NewGithubTeamSupplier(provider, &mockedRepo, deserializer))
		}

		t.Run(c.test, func(tt *testing.T) {
			provider := dritftctlmocks.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.GITHUB), shouldUpdate)
			s := &GithubTeamSupplier{
				provider,
				deserializer,
				&mockedRepo,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)

			mock.AssertExpectationsForObjects(tt)
			test.CtyTestDiff(got, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}
