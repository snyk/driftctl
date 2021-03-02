package github

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/cloudskiff/driftctl/pkg/resource"
	ghdeserializer "github.com/cloudskiff/driftctl/pkg/resource/github/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	dritftctlmocks "github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGithubBranchProtectionSupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *MockGithubRepository)
		err     error
	}{
		{
			test:    "no branch protection",
			dirName: "github_branch_protection_empty",
			mocks: func(client *MockGithubRepository) {
				client.On("ListBranchProtection").Return([]string{}, nil)
			},
			err: nil,
		},
		{
			test:    "Multiple branch protections",
			dirName: "github_branch_protection_multiples",
			mocks: func(client *MockGithubRepository) {
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
	for _, c := range cases {
		shouldUpdate := c.dirName == *goldenfile.Update

		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		mockedRepo := MockGithubRepository{}
		c.mocks(&mockedRepo)

		if shouldUpdate {
			provider, err := InitTestGithubProvider(providerLibrary)
			if err != nil {
				t.Fatal(err)
			}

			supplierLibrary.AddSupplier(NewGithubBranchProtectionSupplier(provider, &mockedRepo))
		}

		t.Run(c.test, func(tt *testing.T) {
			provider := dritftctlmocks.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.GITHUB), shouldUpdate)
			GithubBranchProtectionDeserializer := ghdeserializer.NewGithubBranchProtectionDeserializer()
			s := &GithubBranchProtectionSupplier{
				provider,
				GithubBranchProtectionDeserializer,
				&mockedRepo,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)

			mock.AssertExpectationsForObjects(tt)
			test.CtyTestDiff(got, c.dirName, provider, GithubBranchProtectionDeserializer, shouldUpdate, tt)
		})
	}
}
