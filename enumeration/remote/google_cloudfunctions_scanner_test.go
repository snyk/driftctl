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
	"github.com/snyk/driftctl/mocks"

	testgoogle "github.com/snyk/driftctl/test/google"

	assetpb "cloud.google.com/go/asset/apiv1/assetpb"
	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGoogleCloudFunctionsFunction(t *testing.T) {

	cases := []struct {
		test             string
		assertExpected   func(t *testing.T, got []*resource.Resource)
		response         []*assetpb.Asset
		responseErr      error
		setupAlerterMock func(alerter *mocks.AlerterInterface)
		wantErr          error
	}{
		{
			test:     "no compute instance",
			response: []*assetpb.Asset{},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "one cloud function returned",
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)
				assert.Equal(t, "projects/cloudskiff-dev-elie/locations/us-central1/functions/function-test", got[0].ResourceId())
				assert.Equal(t, "google_cloudfunctions_function", got[0].ResourceType())
			},
			response: []*assetpb.Asset{
				{
					AssetType: "cloudfunctions.googleapis.com/CloudFunction",
					Name:      "//cloudfunctions.googleapis.com/projects/cloudskiff-dev-elie/locations/us-central1/functions/function-test",
					Resource: &assetpb.Resource{
						Data: func() *structpb.Struct {
							v, err := structpb.NewStruct(map[string]interface{}{
								"name": "projects/cloudskiff-dev-elie/locations/us-central1/functions/function-test",
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
			test: "one cloud function without resource data",
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			response: []*assetpb.Asset{
				{
					AssetType: "cloudfunctions.googleapis.com/CloudFunction",
					Name:      "//cloudfunctions.googleapis.com/projects/cloudskiff-dev-elie/locations/us-central1/functions/function-test",
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
					"google_cloudfunctions_function",
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							status.Error(codes.PermissionDenied, "The caller does not have permission"),
							"google_cloudfunctions_function",
						),
						alerts.EnumerationPhase,
					),
				).Once()
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

			assetClient, err := testgoogle.NewFakeAssertServerWithList(c.response, c.responseErr)
			if err != nil {
				tt.Fatal(err)
			}

			realProvider, err := terraform2.InitTestGoogleProvider(providerLibrary, "3.78.0")
			if err != nil {
				tt.Fatal(err)
			}

			repo := repository.NewAssetRepository(assetClient, realProvider.GetConfig(), cache.New(0))

			remoteLibrary.AddEnumerator(google.NewGoogleCloudFunctionsFunctionEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
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
