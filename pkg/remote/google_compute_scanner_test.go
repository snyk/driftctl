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
	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/goldenfile"
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
							status.Error(codes.PermissionDenied, "For scope projects/123456 got error: "+status.Error(codes.PermissionDenied, "The caller does not have permission").Error()+"; "),
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

			repo := repository.NewAssetRepository(assetClient, realProvider.SetConfig([]string{"projects/123456"}), cache.New(0))

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
							status.Error(codes.PermissionDenied, "For scope projects/123456 got error: "+status.Error(codes.PermissionDenied, "The caller does not have permission").Error()+"; "),
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

			repo := repository.NewAssetRepository(assetClient, realProvider.SetConfig([]string{"projects/123456"}), cache.New(0))

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

func TestGoogleComputeInstance(t *testing.T) {

	cases := []struct {
		test             string
		assertExpected   func(t *testing.T, got []*resource.Resource)
		response         []*assetpb.ResourceSearchResult
		responseErr      error
		setupAlerterMock func(alerter *mocks.AlerterInterface)
		wantErr          error
	}{
		{
			test:     "no compute instance",
			response: []*assetpb.ResourceSearchResult{},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiples compute instances",
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)
				assert.Equal(t, "projects/cloudskiff-dev-elie/zones/us-central1-a/instances/test", got[0].ResourceId())
				assert.Equal(t, "google_compute_instance", got[0].ResourceType())
			},
			response: []*assetpb.ResourceSearchResult{
				{
					AssetType: "compute.googleapis.com/Instance",
					Name:      "//compute.googleapis.com/projects/cloudskiff-dev-elie/zones/us-central1-a/instances/test",
				},
			},
		},
		{
			test: "cannot list compute firewall",
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			responseErr: status.Error(codes.PermissionDenied, "The caller does not have permission"),
			setupAlerterMock: func(alerter *mocks.AlerterInterface) {
				alerter.On(
					"SendAlert",
					"google_compute_instance",
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							status.Error(codes.PermissionDenied, "For scope projects/123456 got error: "+status.Error(codes.PermissionDenied, "The caller does not have permission").Error()+"; "),
							"google_compute_instance",
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

			repo := repository.NewAssetRepository(assetClient, realProvider.SetConfig([]string{"projects/123456"}), cache.New(0))

			remoteLibrary.AddEnumerator(google.NewGoogleComputeInstanceEnumerator(repo, factory))

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

func TestGoogleComputeNetwork(t *testing.T) {

	cases := []struct {
		test             string
		dirName          string
		response         []*assetpb.ResourceSearchResult
		responseErr      error
		setupAlerterMock func(alerter *mocks.AlerterInterface)
		wantErr          error
	}{
		{
			test:     "no network",
			dirName:  "google_compute_network_empty",
			response: []*assetpb.ResourceSearchResult{},
			wantErr:  nil,
		},
		{
			test:    "multiple networks",
			dirName: "google_compute_network",
			response: []*assetpb.ResourceSearchResult{
				{
					AssetType:   "compute.googleapis.com/Network",
					DisplayName: "driftctl-unittest-1",
					Name:        "//compute.googleapis.com/projects/driftctl-qa-1/global/networks/driftctl-unittest-1",
				},
				{
					AssetType:   "compute.googleapis.com/Network",
					DisplayName: "driftctl-unittest-2",
					Name:        "//compute.googleapis.com/projects/driftctl-qa-1/global/networks/driftctl-unittest-2",
				},
				{
					AssetType:   "compute.googleapis.com/Network",
					DisplayName: "driftctl-unittest-3",
					Name:        "//compute.googleapis.com/projects/driftctl-qa-1/global/networks/driftctl-unittest-3",
				},
			},
			wantErr: nil,
		},
		{
			test:        "cannot list compute networks",
			dirName:     "google_compute_network_empty",
			responseErr: status.Error(codes.PermissionDenied, "The caller does not have permission"),
			setupAlerterMock: func(alerter *mocks.AlerterInterface) {
				alerter.On(
					"SendAlert",
					"google_compute_network",
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							status.Error(codes.PermissionDenied, "For scope projects/123456 got error: "+status.Error(codes.PermissionDenied, "The caller does not have permission").Error()+"; "),
							"google_compute_network",
						),
						alerts.EnumerationPhase,
					),
				).Once()
			},
			wantErr: nil,
		},
	}

	providerVersion := "3.78.0"
	resType := resource.ResourceType(googleresource.GoogleComputeNetworkResourceType)
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

			repo := repository.NewAssetRepository(assetClient, realProvider.SetConfig([]string{"projects/123456"}), cache.New(0))

			remoteLibrary.AddEnumerator(google.NewGoogleComputeNetworkEnumerator(repo, factory))
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

func TestGoogleComputeInstanceGroup(t *testing.T) {

	cases := []struct {
		test             string
		dirName          string
		response         []*assetpb.ResourceSearchResult
		responseErr      error
		setupAlerterMock func(alerter *mocks.AlerterInterface)
		wantErr          error
	}{
		{
			test:     "no instance group",
			dirName:  "google_compute_instance_group_empty",
			response: []*assetpb.ResourceSearchResult{},
			wantErr:  nil,
		},
		{
			test:    "multiple instance groups",
			dirName: "google_compute_instance_group",
			response: []*assetpb.ResourceSearchResult{
				{
					AssetType:   "compute.googleapis.com/InstanceGroup",
					DisplayName: "driftctl-test-1",
					Name:        "//compute.googleapis.com/projects/cloudskiff-dev-raphael/zones/us-central1-a/instanceGroups/driftctl-test-1",
					Project:     "cloudskiff-dev-raphael",
					Location:    "us-central1-a",
				},
				{
					AssetType:   "compute.googleapis.com/InstanceGroup",
					DisplayName: "driftctl-test-2",
					Name:        "//compute.googleapis.com/projects/cloudskiff-dev-raphael/zones/us-central1-a/instanceGroups/driftctl-test-2",
					Project:     "cloudskiff-dev-raphael",
					Location:    "us-central1-a",
				},
			},
			wantErr: nil,
		},
		{
			test:        "cannot list instance groups",
			dirName:     "google_compute_instance_group_empty",
			responseErr: status.Error(codes.PermissionDenied, "The caller does not have permission"),
			setupAlerterMock: func(alerter *mocks.AlerterInterface) {
				alerter.On(
					"SendAlert",
					"google_compute_instance_group",
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							status.Error(codes.PermissionDenied, "For scope projects/123456 got error: "+status.Error(codes.PermissionDenied, "The caller does not have permission").Error()+"; "),
							"google_compute_instance_group",
						),
						alerts.EnumerationPhase,
					),
				).Once()
			},
			wantErr: nil,
		},
	}

	providerVersion := "3.78.0"
	resType := resource.ResourceType(googleresource.GoogleComputeInstanceGroupResourceType)
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

			repo := repository.NewAssetRepository(assetClient, realProvider.SetConfig([]string{"projects/123456"}), cache.New(0))

			remoteLibrary.AddEnumerator(google.NewGoogleComputeInstanceGroupEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(googleresource.GoogleComputeInstanceGroupResourceType, common.NewGenericDetailsFetcher(googleresource.GoogleComputeInstanceGroupResourceType, provider, deserializer))

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
			test.TestAgainstGoldenFile(got, resType.String(), c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestGoogleComputeAddress(t *testing.T) {

	cases := []struct {
		test             string
		assertExpected   func(t *testing.T, got []*resource.Resource)
		response         []*assetpb.ResourceSearchResult
		responseErr      error
		setupAlerterMock func(alerter *mocks.AlerterInterface)
		wantErr          error
	}{
		{
			test:     "no compute address",
			response: []*assetpb.ResourceSearchResult{},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiples compute address",
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)
				assert.Equal(t, "projects/cloudskiff-dev-elie/regions/us-central1/addresses/my-address", got[0].ResourceId())
				assert.Equal(t, "google_compute_address", got[0].ResourceType())

				assert.Equal(t, "projects/cloudskiff-dev-elie/regions/us-central1/addresses/my-address-2", got[1].ResourceId())
				assert.Equal(t, "google_compute_address", got[1].ResourceType())
				assert.Equal(t, "1.2.3.4", *got[1].Attributes().GetString("address"))
			},
			response: []*assetpb.ResourceSearchResult{
				{
					AssetType: "compute.googleapis.com/Address",
					Name:      "//compute.googleapis.com/projects/cloudskiff-dev-elie/regions/us-central1/addresses/my-address",
				},
				{
					AssetType: "compute.googleapis.com/Address",
					Location:  "global", // Global addresses should be ignored
				},
				{
					AssetType: "compute.googleapis.com/Address",
					Name:      "//compute.googleapis.com/projects/cloudskiff-dev-elie/regions/us-central1/addresses/my-address-2",
					AdditionalAttributes: func() *structpb.Struct {
						str, _ := structpb.NewStruct(map[string]interface{}{
							"address": "1.2.3.4",
						})
						return str
					}(),
				},
			},
		},
		{
			test: "cannot list compute address",
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			responseErr: status.Error(codes.PermissionDenied, "The caller does not have permission"),
			setupAlerterMock: func(alerter *mocks.AlerterInterface) {
				alerter.On(
					"SendAlert",
					"google_compute_address",
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							status.Error(codes.PermissionDenied, "For scope projects/123456 got error: "+status.Error(codes.PermissionDenied, "The caller does not have permission").Error()+"; "),
							"google_compute_address",
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

			assetClient, err := testgoogle.NewFakeAssetServer(c.response, c.responseErr)
			if err != nil {
				tt.Fatal(err)
			}

			realProvider, err := terraform2.InitTestGoogleProvider(providerLibrary, providerVersion)
			if err != nil {
				tt.Fatal(err)
			}

			repo := repository.NewAssetRepository(assetClient, realProvider.SetConfig([]string{"projects/123456"}), cache.New(0))

			remoteLibrary.AddEnumerator(google.NewGoogleComputeAddressEnumerator(repo, factory))

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

func TestGoogleComputeGlobalAddress(t *testing.T) {

	cases := []struct {
		test             string
		assertExpected   func(t *testing.T, got []*resource.Resource)
		response         []*assetpb.Asset
		responseErr      error
		setupAlerterMock func(alerter *mocks.AlerterInterface)
		wantErr          error
	}{
		{
			test:     "no resource",
			response: []*assetpb.Asset{},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "one resource returned",
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)
				assert.Equal(t, "projects/cloudskiff-dev-elie/global/addresses/global-appserver-ip", got[0].ResourceId())
				assert.Equal(t, "google_compute_global_address", got[0].ResourceType())
			},
			response: []*assetpb.Asset{
				{
					AssetType: "compute.googleapis.com/GlobalAddress",
					Name:      "//compute.googleapis.com/projects/cloudskiff-dev-elie/global/addresses/global-appserver-ip",
					Resource: &assetpb.Resource{
						Data: func() *structpb.Struct {
							v, err := structpb.NewStruct(map[string]interface{}{
								"name": "projects/cloudskiff-dev-elie/global/addresses/global-appserver-ip",
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
					AssetType: "compute.googleapis.com/GlobalAddress",
					Name:      "//compute.googleapis.com/projects/cloudskiff-dev-elie/global/addresses/global-appserver-ip",
				},
			},
		},
		{
			test: "cannot list cloud functions",
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			responseErr: status.Error(codes.PermissionDenied, "The caller does not have permission"),
			setupAlerterMock: func(alerter *mocks.AlerterInterface) {
				alerter.On(
					"SendAlert",
					"google_compute_global_address",
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							status.Error(codes.PermissionDenied, "For scope projects/123456 got error: "+status.Error(codes.PermissionDenied, "The caller does not have permission").Error()+"; "),
							"google_compute_global_address",
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

			remoteLibrary.AddEnumerator(google.NewGoogleComputeGlobalAddressEnumerator(repo, factory))

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

func TestGoogleComputeSubnetwork(t *testing.T) {

	cases := []struct {
		test             string
		dirName          string
		response         []*assetpb.ResourceSearchResult
		responseErr      error
		setupAlerterMock func(alerter *mocks.AlerterInterface)
		wantErr          error
	}{
		{
			test:     "no subnetwork",
			dirName:  "google_compute_subnetwork_empty",
			response: []*assetpb.ResourceSearchResult{},
			wantErr:  nil,
		},
		{
			test:    "multiple subnetworks",
			dirName: "google_compute_subnetwork_multiple",
			response: []*assetpb.ResourceSearchResult{
				{
					AssetType:   "compute.googleapis.com/Subnetwork",
					DisplayName: "driftctl-unittest-1",
					Name:        "//compute.googleapis.com/projects/cloudskiff-dev-raphael/regions/us-central1/subnetworks/driftctl-unittest-1",
				},
				{
					AssetType:   "compute.googleapis.com/Subnetwork",
					DisplayName: "driftctl-unittest-2",
					Name:        "//compute.googleapis.com/projects/cloudskiff-dev-raphael/regions/us-central1/subnetworks/driftctl-unittest-2",
				},
				{
					AssetType:   "compute.googleapis.com/Subnetwork",
					DisplayName: "driftctl-unittest-3",
					Name:        "//compute.googleapis.com/projects/cloudskiff-dev-raphael/regions/us-central1/subnetworks/driftctl-unittest-3",
				},
			},
			wantErr: nil,
		},
		{
			test:        "cannot list compute subnetworks",
			dirName:     "google_compute_subnetwork_empty",
			responseErr: status.Error(codes.PermissionDenied, "The caller does not have permission"),
			setupAlerterMock: func(alerter *mocks.AlerterInterface) {
				alerter.On(
					"SendAlert",
					"google_compute_subnetwork",
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							status.Error(codes.PermissionDenied, "For scope projects/123456 got error: "+status.Error(codes.PermissionDenied, "The caller does not have permission").Error()+"; "),
							"google_compute_subnetwork",
						),
						alerts.EnumerationPhase,
					),
				).Once()
			},
			wantErr: nil,
		},
	}

	providerVersion := "3.78.0"
	resType := resource.ResourceType(googleresource.GoogleComputeSubnetworkResourceType)
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

			repo := repository.NewAssetRepository(assetClient, realProvider.SetConfig([]string{"projects/123456"}), cache.New(0))

			remoteLibrary.AddEnumerator(google.NewGoogleComputeSubnetworkEnumerator(repo, factory))
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

func TestGoogleComputeDisk(t *testing.T) {

	cases := []struct {
		test             string
		assertExpected   func(t *testing.T, got []*resource.Resource)
		response         []*assetpb.ResourceSearchResult
		responseErr      error
		setupAlerterMock func(alerter *mocks.AlerterInterface)
		wantErr          error
	}{
		{
			test:     "no compute disk",
			response: []*assetpb.ResourceSearchResult{},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiples compute disk",
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)
				assert.Equal(t, "projects/cloudskiff-dev-elie/zones/us-central1-a/disks/test-disk", got[0].ResourceId())
				assert.Equal(t, "google_compute_disk", got[0].ResourceType())

				assert.Equal(t, "projects/cloudskiff-dev-elie/zones/us-central1-a/disks/test-disk-2", got[1].ResourceId())
				assert.Equal(t, "google_compute_disk", got[1].ResourceType())
			},
			response: []*assetpb.ResourceSearchResult{
				{
					AssetType: "compute.googleapis.com/Disk",
					Name:      "//compute.googleapis.com/projects/cloudskiff-dev-elie/zones/us-central1-a/disks/test-disk",
				},
				{
					AssetType: "compute.googleapis.com/Disk",
					Name:      "//compute.googleapis.com/projects/cloudskiff-dev-elie/zones/us-central1-a/disks/test-disk-2",
				},
			},
		},
		{
			test: "cannot list compute disk",
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			responseErr: status.Error(codes.PermissionDenied, "The caller does not have permission"),
			setupAlerterMock: func(alerter *mocks.AlerterInterface) {
				alerter.On(
					"SendAlert",
					"google_compute_disk",
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							status.Error(codes.PermissionDenied, "For scope projects/123456 got error: "+status.Error(codes.PermissionDenied, "The caller does not have permission").Error()+"; "),
							"google_compute_disk",
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

			assetClient, err := testgoogle.NewFakeAssetServer(c.response, c.responseErr)
			if err != nil {
				tt.Fatal(err)
			}

			realProvider, err := terraform2.InitTestGoogleProvider(providerLibrary, providerVersion)
			if err != nil {
				tt.Fatal(err)
			}

			repo := repository.NewAssetRepository(assetClient, realProvider.SetConfig([]string{"projects/123456"}), cache.New(0))

			remoteLibrary.AddEnumerator(google.NewGoogleComputeDiskEnumerator(repo, factory))

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

func TestGoogleComputeImage(t *testing.T) {

	cases := []struct {
		test             string
		assertExpected   func(t *testing.T, got []*resource.Resource)
		response         []*assetpb.ResourceSearchResult
		responseErr      error
		setupAlerterMock func(alerter *mocks.AlerterInterface)
		wantErr          error
	}{
		{
			test:     "no compute image",
			response: []*assetpb.ResourceSearchResult{},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiples images",
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)
				assert.Equal(t, "projects/cloudskiff-dev-elie/global/images/example-image", got[0].ResourceId())
				assert.Equal(t, "google_compute_image", got[0].ResourceType())

				assert.Equal(t, "projects/cloudskiff-dev-elie/global/images/example-image-2", got[1].ResourceId())
				assert.Equal(t, "google_compute_image", got[1].ResourceType())
			},
			response: []*assetpb.ResourceSearchResult{
				{
					AssetType: "compute.googleapis.com/Image",
					Name:      "//compute.googleapis.com/projects/cloudskiff-dev-elie/global/images/example-image",
				},
				{
					AssetType: "compute.googleapis.com/Image",
					Name:      "//compute.googleapis.com/projects/cloudskiff-dev-elie/global/images/example-image-2",
				},
			},
		},
		{
			test: "cannot list images",
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			responseErr: status.Error(codes.PermissionDenied, "The caller does not have permission"),
			setupAlerterMock: func(alerter *mocks.AlerterInterface) {
				alerter.On(
					"SendAlert",
					"google_compute_image",
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							status.Error(codes.PermissionDenied, "For scope projects/123456 got error: "+status.Error(codes.PermissionDenied, "The caller does not have permission").Error()+"; "),
							"google_compute_image",
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

			assetClient, err := testgoogle.NewFakeAssetServer(c.response, c.responseErr)
			if err != nil {
				tt.Fatal(err)
			}

			realProvider, err := terraform2.InitTestGoogleProvider(providerLibrary, providerVersion)
			if err != nil {
				tt.Fatal(err)
			}

			repo := repository.NewAssetRepository(assetClient, realProvider.SetConfig([]string{"projects/123456"}), cache.New(0))

			remoteLibrary.AddEnumerator(google.NewGoogleComputeImageEnumerator(repo, factory))

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

func TestGoogleComputeHealthCheck(t *testing.T) {

	cases := []struct {
		test             string
		assertExpected   func(t *testing.T, got []*resource.Resource)
		response         []*assetpb.ResourceSearchResult
		responseErr      error
		setupAlerterMock func(alerter *mocks.AlerterInterface)
		wantErr          error
	}{
		{
			test:     "no compute health check",
			response: []*assetpb.ResourceSearchResult{},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiples compute health checks",
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)
				assert.Equal(t, "projects/cloudskiff-dev-raphael/global/healthChecks/test-health-check-1", got[0].ResourceId())
				assert.Equal(t, "google_compute_health_check", got[0].ResourceType())

				assert.Equal(t, "projects/cloudskiff-dev-raphael/global/healthChecks/test-health-check-2", got[1].ResourceId())
				assert.Equal(t, "google_compute_health_check", got[1].ResourceType())
			},
			response: []*assetpb.ResourceSearchResult{
				{
					AssetType: "compute.googleapis.com/HealthCheck",
					Name:      "//compute.googleapis.com/projects/cloudskiff-dev-raphael/global/healthChecks/test-health-check-1",
				},
				{
					AssetType: "compute.googleapis.com/HealthCheck",
					Name:      "//compute.googleapis.com/projects/cloudskiff-dev-raphael/global/healthChecks/test-health-check-2",
				},
			},
		},
		{
			test: "cannot list compute health checks",
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			responseErr: status.Error(codes.PermissionDenied, "The caller does not have permission"),
			setupAlerterMock: func(alerter *mocks.AlerterInterface) {
				alerter.On(
					"SendAlert",
					"google_compute_health_check",
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							status.Error(codes.PermissionDenied, "For scope projects/123456 got error: "+status.Error(codes.PermissionDenied, "The caller does not have permission").Error()+"; "),
							"google_compute_health_check",
						),
						alerts.EnumerationPhase,
					),
				).Once()
			},
		},
	}

	providerVersion := "3.78.0"
	schemaRepository := testresource.InitFakeSchemaRepository(terraform.GOOGLE, providerVersion)
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

			repo := repository.NewAssetRepository(assetClient, realProvider.SetConfig([]string{"projects/123456"}), cache.New(0))

			remoteLibrary.AddEnumerator(google.NewGoogleComputeHealthCheckEnumerator(repo, factory))

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

func TestGoogleComputeNodeGroup(t *testing.T) {

	cases := []struct {
		test             string
		assertExpected   func(t *testing.T, got []*resource.Resource)
		response         []*assetpb.Asset
		responseErr      error
		setupAlerterMock func(alerter *mocks.AlerterInterface)
		wantErr          error
	}{
		{
			test:     "no compute node group",
			response: []*assetpb.Asset{},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiples compute node group",
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)
				assert.Equal(t, "projects/cloudskiff-dev-martin/zones/us-central1-f/nodeGroups/soletenant-group", got[0].ResourceId())
				assert.Equal(t, "google_compute_node_group", got[0].ResourceType())

				assert.Equal(t, "projects/cloudskiff-dev-martin/zones/us-central1-f/nodeGroups/simple-group", got[1].ResourceId())
				assert.Equal(t, "google_compute_node_group", got[1].ResourceType())
			},
			response: []*assetpb.Asset{
				{
					AssetType: "compute.googleapis.com/NodeGroup",
					Name:      "//compute.googleapis.com/projects/cloudskiff-dev-martin/zones/us-central1-f/nodeGroups/soletenant-group",
				},
				{
					AssetType: "compute.googleapis.com/NodeGroup",
					Name:      "//compute.googleapis.com/projects/cloudskiff-dev-martin/zones/us-central1-f/nodeGroups/simple-group",
				},
			},
		},
		{
			test: "cannot list compute node group",
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			responseErr: status.Error(codes.PermissionDenied, "The caller does not have permission"),
			setupAlerterMock: func(alerter *mocks.AlerterInterface) {
				alerter.On(
					"SendAlert",
					"google_compute_node_group",
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							status.Error(codes.PermissionDenied, "For scope projects/123456 got error: "+status.Error(codes.PermissionDenied, "The caller does not have permission").Error()+"; "),
							"google_compute_node_group",
						),
						alerts.EnumerationPhase,
					),
				).Once()
			},
		},
	}

	providerVersion := "3.78.0"
	schemaRepository := testresource.InitFakeSchemaRepository(terraform.GOOGLE, providerVersion)
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

			remoteLibrary.AddEnumerator(google.NewGoogleComputeNodeGroupEnumerator(repo, factory))

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
