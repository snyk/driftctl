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
)

func TestGoogleDNSNanagedZone(t *testing.T) {

	cases := []struct {
		test             string
		response         []*assetpb.ResourceSearchResult
		responseErr      error
		setupAlerterMock func(alerter *mocks.AlerterInterface)
		wantErr          error
		assertExpected   func(t *testing.T, got []*resource.Resource)
	}{
		{
			test:     "no managed zone",
			response: []*assetpb.ResourceSearchResult{},
			wantErr:  nil,
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiples managed zones",
			response: []*assetpb.ResourceSearchResult{
				{
					AssetType: "dns.googleapis.com/ManagedZone",
					Name:      "invalid ID", // Should be ignored
				},
				{
					AssetType:   "dns.googleapis.com/ManagedZone",
					DisplayName: "test-zone-0",
					Name:        "//dns.googleapis.com/projects/cloudskiff-dev-raphael/managedZones/123456789",
				},
				{
					AssetType:   "dns.googleapis.com/ManagedZone",
					DisplayName: "test-zone-1",
					Name:        "//dns.googleapis.com/projects/cloudskiff-dev-raphael/managedZones/123456789",
				},
				{
					AssetType:   "dns.googleapis.com/ManagedZone",
					DisplayName: "test-zone-2",
					Name:        "//dns.googleapis.com/projects/cloudskiff-dev-raphael/managedZones/123456789",
				},
			},
			wantErr: nil,
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 3)

				assert.Equal(t, got[0].ResourceId(), "projects/cloudskiff-dev-raphael/managedZones/test-zone-0")
				assert.Equal(t, got[0].ResourceType(), googleresource.GoogleDNSManagedZoneResourceType)

				assert.Equal(t, got[1].ResourceId(), "projects/cloudskiff-dev-raphael/managedZones/test-zone-1")
				assert.Equal(t, got[1].ResourceType(), googleresource.GoogleDNSManagedZoneResourceType)

				assert.Equal(t, got[2].ResourceId(), "projects/cloudskiff-dev-raphael/managedZones/test-zone-2")
				assert.Equal(t, got[2].ResourceType(), googleresource.GoogleDNSManagedZoneResourceType)
			},
		},
		{
			test:        "should return access denied error",
			wantErr:     nil,
			responseErr: status.Error(codes.PermissionDenied, "The caller does not have permission"),
			setupAlerterMock: func(alerter *mocks.AlerterInterface) {
				alerter.On(
					"SendAlert",
					googleresource.GoogleDNSManagedZoneResourceType,
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							status.Error(codes.PermissionDenied, "The caller does not have permission"),
							googleresource.GoogleDNSManagedZoneResourceType,
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

			remoteLibrary.AddEnumerator(google.NewGoogleDNSManagedZoneEnumerator(repo, factory))

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
