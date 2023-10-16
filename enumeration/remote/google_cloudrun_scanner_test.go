package remote

import (
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/remote/google"
	"github.com/snyk/driftctl/enumeration/remote/google/repository"
	"github.com/snyk/driftctl/enumeration/terraform"

	"github.com/snyk/driftctl/enumeration/resource"
	googleresource "github.com/snyk/driftctl/enumeration/resource/google"
	"github.com/snyk/driftctl/mocks"

	testgoogle "github.com/snyk/driftctl/test/google"

	assetpb "cloud.google.com/go/asset/apiv1/assetpb"
	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGoogleCloudRunService(t *testing.T) {

	cases := []struct {
		test             string
		response         []*assetpb.ResourceSearchResult
		responseErr      error
		setupAlerterMock func(alerter *mocks.AlerterInterface)
		wantErr          error
		assertExpected   func(t *testing.T, got []*resource.Resource)
	}{
		{
			test:     "no resource",
			response: []*assetpb.ResourceSearchResult{},
			wantErr:  nil,
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiples resources",
			response: []*assetpb.ResourceSearchResult{
				{
					AssetType: "run.googleapis.com/Service",
					Name:      "invalid ID", // Should be ignored
				},
				{
					AssetType:   "run.googleapis.com/Service",
					DisplayName: "cloudrun-srv-1",
					Name:        "//run.googleapis.com/projects/cloudskiff-dev-elie/locations/us-central1/services/cloudrun-srv-1",
					Location:    "us-central1",
				},
				{
					AssetType:   "run.googleapis.com/Service",
					DisplayName: "cloudrun-srv-2",
					Name:        "//run.googleapis.com/projects/cloudskiff-dev-elie/locations/us-central1/services/cloudrun-srv-2",
					Location:    "us-central1",
				},
			},
			wantErr: nil,
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "locations/us-central1/namespaces/cloudskiff-dev-elie/services/cloudrun-srv-1")
				assert.Equal(t, got[0].ResourceType(), googleresource.GoogleCloudRunServiceResourceType)

				assert.Equal(t, got[1].ResourceId(), "locations/us-central1/namespaces/cloudskiff-dev-elie/services/cloudrun-srv-2")
				assert.Equal(t, got[1].ResourceType(), googleresource.GoogleCloudRunServiceResourceType)
			},
		},
		{
			test:        "should return access denied error",
			wantErr:     nil,
			responseErr: status.Error(codes.PermissionDenied, "The caller does not have permission"),
			setupAlerterMock: func(alerter *mocks.AlerterInterface) {
				alerter.On(
					"SendAlert",
					googleresource.GoogleCloudRunServiceResourceType,
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							status.Error(codes.PermissionDenied, "The caller does not have permission"),
							googleresource.GoogleCloudRunServiceResourceType,
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

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
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

			realProvider, err := terraform2.InitTestGoogleProvider(providerLibrary, "3.78.0")
			if err != nil {
				tt.Fatal(err)
			}

			repo := repository.NewAssetRepository(assetClient, realProvider.GetConfig(), cache.New(0))

			remoteLibrary.AddEnumerator(google.NewGoogleCloudRunServiceEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)
			if err != nil {
				return
			}

			alerter.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
			if c.assertExpected != nil {
				c.assertExpected(tt, got)
			}
		})
	}
}
