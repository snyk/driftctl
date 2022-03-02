package remote

import (
	"testing"

	"github.com/snyk/driftctl/mocks"
	"github.com/snyk/driftctl/pkg/filter"
	"github.com/snyk/driftctl/pkg/remote/alerts"
	"github.com/snyk/driftctl/pkg/remote/cache"
	"github.com/snyk/driftctl/pkg/remote/common"
	remoteerr "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/remote/google"
	"github.com/snyk/driftctl/pkg/remote/google/repository"
	"github.com/snyk/driftctl/pkg/resource"
	googleresource "github.com/snyk/driftctl/pkg/resource/google"
	"github.com/snyk/driftctl/pkg/terraform"
	testgoogle "github.com/snyk/driftctl/test/google"
	testresource "github.com/snyk/driftctl/test/resource"
	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	assetpb "google.golang.org/genproto/googleapis/cloud/asset/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGoogleSQLDatabaseInstance(t *testing.T) {

	cases := []struct {
		test             string
		assertExpected   func(t *testing.T, got []*resource.Resource)
		response         []*assetpb.Asset
		responseErr      error
		setupAlerterMock func(alerter *mocks.AlerterInterface)
		wantErr          error
	}{
		{
			test:     "no instance",
			response: []*assetpb.Asset{},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "one resource returned",
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)
				assert.Equal(t, "instance-test", got[0].ResourceId())
				assert.Equal(t, "google_sql_database_instance", got[0].ResourceType())
			},
			response: []*assetpb.Asset{
				{
					AssetType: "sqladmin.googleapis.com/Instance",
					Resource: &assetpb.Resource{
						Data: func() *structpb.Struct {
							v, err := structpb.NewStruct(map[string]interface{}{
								"name": "instance-test",
							})
							if err != nil {
								t.Fatal(err)
							}
							return v
						}(),
					},
				},
			},
		},
		{
			test: "one resource without resource data",
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			response: []*assetpb.Asset{
				{
					AssetType: "sqladmin.googleapis.com/Instance",
				},
			},
		},
		{
			test: "cannot list resources",
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			responseErr: status.Error(codes.PermissionDenied, "The caller does not have permission"),
			setupAlerterMock: func(alerter *mocks.AlerterInterface) {
				alerter.On(
					"SendAlert",
					"google_sql_database_instance",
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							status.Error(codes.PermissionDenied, "For scope projects/123456 got error: "+status.Error(codes.PermissionDenied, "The caller does not have permission").Error()+"; "),
							"google_sql_database_instance",
						),
						alerts.EnumerationPhase,
					),
				).Once()
			},
		},
	}

	providerVersion := "3.78.0"
	schemaRepository := testresource.InitFakeSchemaRepository("google", providerVersion)
	googleresource.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			scanOptions := ScannerOptions{}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			if c.setupAlerterMock != nil {
				c.setupAlerterMock(alerter)
			}

			assetClient, err := testgoogle.NewFakeAssertServerWithList(c.response, c.responseErr)
			if err != nil {
				tt.Fatal(err)
			}

			realProvider, err := terraform2.InitTestGoogleProvider(providerLibrary, providerVersion)
			if err != nil {
				tt.Fatal(err)
			}

			repo := repository.NewAssetRepository(assetClient, realProvider.SetConfig([]string{"projects/123456"}), cache.New(0))

			remoteLibrary.AddEnumerator(google.NewGoogleSQLDatabaseInstanceEnumerator(repo, factory))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			alerter.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
			if c.assertExpected != nil {
				c.assertExpected(t, got)
			}
		})
	}
}
