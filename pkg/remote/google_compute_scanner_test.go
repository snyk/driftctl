package remote

import (
	"testing"

	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/remote/alerts"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	remoteerr "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/remote/google"
	"github.com/cloudskiff/driftctl/pkg/remote/google/repository"
	"github.com/cloudskiff/driftctl/pkg/resource"
	googleresource "github.com/cloudskiff/driftctl/pkg/resource/google"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	testgoogle "github.com/cloudskiff/driftctl/test/google"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	terraform2 "github.com/cloudskiff/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	assetpb "google.golang.org/genproto/googleapis/cloud/asset/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGoogleComputeFirewall(t *testing.T) {

	cases := []struct {
		test             string
		dirName          string
		response         []*assetpb.ResourceSearchResult
		responseErr      error
		setupAlerterMock func(alerter *mocks.AlerterInterface)
		wantErr          error
	}{
		{
			test:     "no compute firewall",
			dirName:  "google_compute_firewall_empty",
			response: []*assetpb.ResourceSearchResult{},
			wantErr:  nil,
		},
		{
			test:    "multiples compute firewall",
			dirName: "google_compute_firewall",
			response: []*assetpb.ResourceSearchResult{
				{
					AssetType:   "compute.googleapis.com/Firewall",
					DisplayName: "test-firewall-0",
					Name:        "//compute.googleapis.com/projects/cloudskiff-dev-elie/global/firewalls/test-firewall-0",
				},
				{
					AssetType:   "compute.googleapis.com/Firewall",
					DisplayName: "test-firewall-1",
					Name:        "//compute.googleapis.com/projects/cloudskiff-dev-elie/global/firewalls/test-firewall-1",
				},
				{
					AssetType:   "compute.googleapis.com/Firewall",
					DisplayName: "test-firewall-2",
					Name:        "//compute.googleapis.com/projects/cloudskiff-dev-elie/global/firewalls/test-firewall-2",
				},
			},
			wantErr: nil,
		},
		{
			test:        "cannot list compute firewall",
			dirName:     "google_compute_firewall_empty",
			responseErr: status.Error(codes.PermissionDenied, "The caller does not have permission"),
			setupAlerterMock: func(alerter *mocks.AlerterInterface) {
				alerter.On(
					"SendAlert",
					"google_compute_firewall",
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							status.Error(codes.PermissionDenied, "The caller does not have permission"),
							"google_compute_firewall",
						),
						alerts.EnumerationPhase,
					),
				).Once()
			},
			wantErr: nil,
		},
	}

	providerVersion := "3.78.0"
	resType := resource.ResourceType(googleresource.GoogleComputeFirewallResourceType)
	schemaRepository := testresource.InitFakeSchemaRepository("google", providerVersion)
	googleresource.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			if c.setupAlerterMock != nil {
				c.setupAlerterMock(alerter)
			}

			assetClient, err := testgoogle.NewFakeAssetServer(c.response, c.responseErr)
			if err != nil {
				tt.Fatal(err)
			}

			realProvider, err := terraform2.InitTestGoogleProvider(providerLibrary, providerVersion)
			if err != nil {
				tt.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err = realProvider.Init()
				if err != nil {
					tt.Fatal(err)
				}
				provider.ShouldUpdate()
			}

			repo := repository.NewAssetRepository(assetClient, realProvider.GetConfig(), cache.New(0))

			remoteLibrary.AddEnumerator(google.NewGoogleComputeFirewallEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resType, common.NewGenericDetailsFetcher(resType, provider, deserializer))

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
			test.TestAgainstGoldenFile(got, resType.String(), c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestGoogleComputeRouter(t *testing.T) {

	cases := []struct {
		test             string
		response         []*assetpb.ResourceSearchResult
		responseErr      error
		setupAlerterMock func(alerter *mocks.AlerterInterface)
		wantErr          error
		assertExpected   func(t *testing.T, got []*resource.Resource)
	}{
		{
			test:     "no compute router",
			response: []*assetpb.ResourceSearchResult{},
			wantErr:  nil,
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiples compute routers",
			response: []*assetpb.ResourceSearchResult{
				{
					AssetType:   "compute.googleapis.com/Router",
					DisplayName: "test-router-0",
					Name:        "//compute.googleapis.com/projects/cloudskiff-dev-raphael/regions/us-central1/routers/test-router-0",
				},
				{
					AssetType:   "compute.googleapis.com/Router",
					DisplayName: "test-router-1",
					Name:        "//compute.googleapis.com/projects/cloudskiff-dev-raphael/regions/us-central1/routers/test-router-1",
				},
				{
					AssetType:   "compute.googleapis.com/Router",
					DisplayName: "test-router-2",
					Name:        "//compute.googleapis.com/projects/cloudskiff-dev-raphael/regions/us-central1/routers/test-router-2",
				},
			},
			wantErr: nil,
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 3)

				assert.Equal(t, got[0].ResourceId(), "projects/cloudskiff-dev-raphael/regions/us-central1/routers/test-router-0")
				assert.Equal(t, got[0].ResourceType(), googleresource.GoogleComputeRouterResourceType)

				assert.Equal(t, got[1].ResourceId(), "projects/cloudskiff-dev-raphael/regions/us-central1/routers/test-router-1")
				assert.Equal(t, got[1].ResourceType(), googleresource.GoogleComputeRouterResourceType)

				assert.Equal(t, got[2].ResourceId(), "projects/cloudskiff-dev-raphael/regions/us-central1/routers/test-router-2")
				assert.Equal(t, got[2].ResourceType(), googleresource.GoogleComputeRouterResourceType)
			},
		},
		{
			test:        "should return access denied error",
			wantErr:     nil,
			responseErr: status.Error(codes.PermissionDenied, "The caller does not have permission"),
			setupAlerterMock: func(alerter *mocks.AlerterInterface) {
				alerter.On(
					"SendAlert",
					googleresource.GoogleComputeRouterResourceType,
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							status.Error(codes.PermissionDenied, "The caller does not have permission"),
							googleresource.GoogleComputeRouterResourceType,
						),
						alerts.EnumerationPhase,
					),
				).Once()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
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

			assetClient, err := testgoogle.NewFakeAssetServer(c.response, c.responseErr)
			if err != nil {
				tt.Fatal(err)
			}

			realProvider, err := terraform2.InitTestGoogleProvider(providerLibrary, providerVersion)
			if err != nil {
				tt.Fatal(err)
			}

			repo := repository.NewAssetRepository(assetClient, realProvider.GetConfig(), cache.New(0))

			remoteLibrary.AddEnumerator(google.NewGoogleComputeRouterEnumerator(repo, factory))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)
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
