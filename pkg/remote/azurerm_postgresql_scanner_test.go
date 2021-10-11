package remote

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/postgresql/armpostgresql"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm"
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	error2 "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceazure "github.com/cloudskiff/driftctl/pkg/resource/azurerm"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAzurermPostgresqlServer(t *testing.T) {

	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockPostgresqlRespository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no postgres server",
			mocks: func(repository *repository.MockPostgresqlRespository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllServers").Return([]*armpostgresql.Server{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "error listing postgres servers",
			mocks: func(repository *repository.MockPostgresqlRespository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllServers").Return(nil, dummyError)
			},
			wantErr: error2.NewResourceListingError(dummyError, resourceazure.AzurePostgresqlServerResourceType),
		},
		{
			test: "multiple postgres servers",
			mocks: func(repository *repository.MockPostgresqlRespository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllServers").Return([]*armpostgresql.Server{
					{
						TrackedResource: armpostgresql.TrackedResource{
							Resource: armpostgresql.Resource{
								ID:   to.StringPtr("server1"),
								Name: to.StringPtr("server1"),
							},
						},
					},
					{
						TrackedResource: armpostgresql.TrackedResource{
							Resource: armpostgresql.Resource{
								ID:   to.StringPtr("server2"),
								Name: to.StringPtr("server2"),
							},
						},
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "server1")
				assert.Equal(t, got[0].ResourceType(), resourceazure.AzurePostgresqlServerResourceType)

				assert.Equal(t, got[1].ResourceId(), "server2")
				assert.Equal(t, got[1].ResourceType(), resourceazure.AzurePostgresqlServerResourceType)
			},
		},
	}

	providerVersion := "2.71.0"
	schemaRepository := testresource.InitFakeSchemaRepository("azurerm", providerVersion)
	resourceazure.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {

			scanOptions := ScannerOptions{}
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockPostgresqlRespository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.PostgresqlRespository = fakeRepo

			remoteLibrary.AddEnumerator(azurerm.NewAzurermPostgresqlServerEnumerator(repo, factory))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
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
