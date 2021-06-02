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

func TestGithubTeamMembershipSupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *MockGithubRepository)
		err     error
	}{
		{
			test:    "no github team memberships",
			dirName: "github_team_membership_empty",
			mocks: func(client *MockGithubRepository) {
				client.On("ListTeamMemberships").Return([]string{}, nil)
			},
			err: nil,
		},
		{
			test:    "multiple github team memberships",
			dirName: "github_team_membership_multiple",
			mocks: func(client *MockGithubRepository) {
				client.On("ListTeamMemberships").Return([]string{
					"4570529:driftctl-acceptance-tester",
					"4570529:wbeuil",
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

			supplierLibrary.AddSupplier(NewGithubTeamMembershipSupplier(provider, &mockedRepo, deserializer))
		}

		t.Run(c.test, func(tt *testing.T) {
			provider := dritftctlmocks.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.GITHUB), shouldUpdate)
			s := &GithubTeamMembershipSupplier{
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
