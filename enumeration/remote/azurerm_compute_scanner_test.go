package remote

import (
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/azurerm"
	"github.com/snyk/driftctl/enumeration/remote/azurerm/repository"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/terraform"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration/resource"
	resourceazure "github.com/snyk/driftctl/enumeration/resource/azurerm"
	"github.com/snyk/driftctl/mocks"

	"github.com/snyk/driftctl/test/goldenfile"
	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAzurermCompute_Image(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockComputeRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no images",
			mocks: func(repository *repository.MockComputeRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllImages").Return([]*armcompute.Image{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "error listing images",
			mocks: func(repository *repository.MockComputeRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllImages").Return(nil, dummyError)
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceazure.AzureImageResourceType),
		},
		{
			test: "multiple images including an invalid ID",
			mocks: func(repository *repository.MockComputeRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllImages").Return([]*armcompute.Image{
					{
						Resource: armcompute.Resource{
							ID:   to.StringPtr("/subscriptions/4e411884-65b0-4911-bc80-52f9a21942a2/resourceGroups/testgroup/providers/Microsoft.Compute/images/image1"),
							Name: to.StringPtr("image1"),
						},
					},
					{
						Resource: armcompute.Resource{
							ID:   to.StringPtr("/subscriptions/4e411884-65b0-4911-bc80-52f9a21942a2/resourceGroups/testgroup/providers/Microsoft.Compute/images/image2"),
							Name: to.StringPtr("image2"),
						},
					},
					{
						Resource: armcompute.Resource{
							ID:   to.StringPtr("/invalid-id/image3"),
							Name: to.StringPtr("image3"),
						},
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "/subscriptions/4e411884-65b0-4911-bc80-52f9a21942a2/resourceGroups/testgroup/providers/Microsoft.Compute/images/image1")
				assert.Equal(t, got[0].ResourceType(), resourceazure.AzureImageResourceType)

				assert.Equal(t, got[1].ResourceId(), "/subscriptions/4e411884-65b0-4911-bc80-52f9a21942a2/resourceGroups/testgroup/providers/Microsoft.Compute/images/image2")
				assert.Equal(t, got[1].ResourceType(), resourceazure.AzureImageResourceType)
			},
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockComputeRepository{}
			c.mocks(fakeRepo, alerter)

			remoteLibrary.AddEnumerator(azurerm.NewAzurermImageEnumerator(fakeRepo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestAzurermCompute_SSHPublicKey(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		dirName        string
		mocks          func(*repository.MockComputeRepository, *mocks.AlerterInterface)
		assertExpected func(*testing.T, []*resource.Resource)
		wantErr        error
	}{
		{
			test:    "no public key",
			dirName: "azurerm_ssh_public_key_empty",
			mocks: func(repository *repository.MockComputeRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllSSHPublicKeys").Return([]*armcompute.SSHPublicKeyResource{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test:    "error listing public keys",
			dirName: "azurerm_ssh_public_key_empty",
			mocks: func(repository *repository.MockComputeRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllSSHPublicKeys").Return(nil, dummyError)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceazure.AzureSSHPublicKeyResourceType),
		},
		{
			test:    "multiple public keys",
			dirName: "azurerm_ssh_public_key_multiple",
			mocks: func(repository *repository.MockComputeRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllSSHPublicKeys").Return([]*armcompute.SSHPublicKeyResource{
					{
						Resource: armcompute.Resource{
							ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/TESTRESGROUP/providers/Microsoft.Compute/sshPublicKeys/example-key"),
							Name: to.StringPtr("example-key"),
						},
					},
					{
						Resource: armcompute.Resource{
							ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/TESTRESGROUP/providers/Microsoft.Compute/sshPublicKeys/example-key2"),
							Name: to.StringPtr("example-key2"),
						},
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, "/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/TESTRESGROUP/providers/Microsoft.Compute/sshPublicKeys/example-key", got[0].ResourceId())
				assert.Equal(t, resourceazure.AzureSSHPublicKeyResourceType, got[0].ResourceType())

				assert.Equal(t, "/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/TESTRESGROUP/providers/Microsoft.Compute/sshPublicKeys/example-key2", got[1].ResourceId())
				assert.Equal(t, resourceazure.AzureSSHPublicKeyResourceType, got[1].ResourceType())
			},
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockComputeRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ComputeRepository = fakeRepo
			providerVersion := "2.71.0"
			realProvider, err := terraform2.InitTestAzureProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{})
				if err != nil {
					t.Fatal(err)
				}
				clientOptions := &arm.ClientOptions{}
				repo = repository.NewComputeRepository(cred, clientOptions, realProvider.GetConfig(), cache.New(0))
			}

			remoteLibrary.AddEnumerator(azurerm.NewAzurermSSHPublicKeyEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)

			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}
