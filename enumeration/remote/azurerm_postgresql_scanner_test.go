package remote

import (
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/azurerm"
	"github.com/snyk/driftctl/enumeration/remote/azurerm/repository"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/terraform"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration/resource"
	resourceazure "github.com/snyk/driftctl/enumeration/resource/azurerm"
	"github.com/snyk/driftctl/mocks"

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
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceazure.AzurePostgresqlServerResourceType),
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

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockPostgresqlRespository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.PostgresqlRespository = fakeRepo

			remoteLibrary.AddEnumerator(azurerm.NewAzurermPostgresqlServerEnumerator(repo, factory))

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

func TestAzurermPostgresqlDatabase(t *testing.T) {

	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockPostgresqlRespository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no postgres database",
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
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceazure.AzurePostgresqlDatabaseResourceType, resourceazure.AzurePostgresqlServerResourceType),
		},
		{
			test: "error listing postgres databases",
			mocks: func(repository *repository.MockPostgresqlRespository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllServers").Return([]*armpostgresql.Server{
					{
						TrackedResource: armpostgresql.TrackedResource{
							Resource: armpostgresql.Resource{
								ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/api-rg-pro/providers/Microsoft.DBforPostgreSQL/servers/postgresql-server-8791542"),
								Name: to.StringPtr("postgresql-server-8791542"),
							},
						},
					},
				}, nil).Once()

				repository.On("ListAllDatabasesByServer", mock.IsType(&armpostgresql.Server{})).Return(nil, dummyError).Once()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceazure.AzurePostgresqlDatabaseResourceType),
		},
		{
			test: "multiple postgres databases",
			mocks: func(repository *repository.MockPostgresqlRespository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllServers").Return([]*armpostgresql.Server{
					{
						TrackedResource: armpostgresql.TrackedResource{
							Resource: armpostgresql.Resource{
								ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/api-rg-pro/providers/Microsoft.DBforPostgreSQL/servers/postgresql-server-8791542"),
								Name: to.StringPtr("postgresql-server-8791542"),
							},
						},
					},
				}, nil).Once()

				repository.On("ListAllDatabasesByServer", mock.IsType(&armpostgresql.Server{})).Return([]*armpostgresql.Database{
					{
						ProxyResource: armpostgresql.ProxyResource{
							Resource: armpostgresql.Resource{
								ID:   to.StringPtr("db1"),
								Name: to.StringPtr("db1"),
							},
						},
					},
					{
						ProxyResource: armpostgresql.ProxyResource{
							Resource: armpostgresql.Resource{
								ID:   to.StringPtr("db2"),
								Name: to.StringPtr("db2"),
							},
						},
					},
				}, nil).Once()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "db1")
				assert.Equal(t, got[0].ResourceType(), resourceazure.AzurePostgresqlDatabaseResourceType)

				assert.Equal(t, got[1].ResourceId(), "db2")
				assert.Equal(t, got[1].ResourceType(), resourceazure.AzurePostgresqlDatabaseResourceType)
			},
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockPostgresqlRespository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.PostgresqlRespository = fakeRepo

			remoteLibrary.AddEnumerator(azurerm.NewAzurermPostgresqlDatabaseEnumerator(repo, factory))

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
