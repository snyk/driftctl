package remote

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/snyk/driftctl/mocks"
	"github.com/snyk/driftctl/pkg/filter"
	"github.com/snyk/driftctl/pkg/remote/alerts"
	"github.com/snyk/driftctl/pkg/remote/common"
	remoteerr "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/remote/google"
	"github.com/snyk/driftctl/pkg/remote/google/repository"
	"github.com/snyk/driftctl/pkg/resource"
	googleresource "github.com/snyk/driftctl/pkg/resource/google"
	"github.com/snyk/driftctl/pkg/terraform"
	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/goldenfile"
	testresource "github.com/snyk/driftctl/test/resource"
	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGoogleProjectIAMMember(t *testing.T) {

	cases := []struct {
		test             string
		dirName          string
		repositoryMock   func(repository *repository.MockCloudResourceManagerRepository)
		responseErr      error
		setupAlerterMock func(alerter *mocks.AlerterInterface)
		wantErr          error
	}{
		{
			test:    "no bindings",
			dirName: "google_project_member_empty",
			repositoryMock: func(repository *repository.MockCloudResourceManagerRepository) {
				repository.On("ListProjectsBindings").Return(map[string]map[string][]string{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "Cannot list bindings",
			dirName: "google_project_member_listing_error",
			repositoryMock: func(repository *repository.MockCloudResourceManagerRepository) {
				repository.On("ListProjectsBindings").Return(
					map[string]map[string][]string{},
					errors.New("googleapi: Error 403: driftctl-acc-circle@driftctl-qa-1.iam.gserviceaccount.com does not have project.getIamPolicy access., forbidden"))
			},
			setupAlerterMock: func(alerter *mocks.AlerterInterface) {
				alerter.On(
					"SendAlert",
					"google_project_iam_member",
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							errors.New("googleapi: Error 403: driftctl-acc-circle@driftctl-qa-1.iam.gserviceaccount.com does not have project.getIamPolicy access., forbidden"),
							"google_project_iam_member",
						),
						alerts.EnumerationPhase,
					),
				).Once()
			},
			wantErr: nil,
		},
		{
			test:    "multiples storage buckets, multiple bindings",
			dirName: "google_project_member_listing_multiple",
			repositoryMock: func(repository *repository.MockCloudResourceManagerRepository) {
				repository.On("ListProjectsBindings").Return(map[string]map[string][]string{
					"": {
						"roles/editor": {
							"user:martin.guibert@cloudskiff.com",
							"serviceAccount:drifctl-admin@cloudskiff-dev-martin.iam.gserviceaccount.com",
						},
						"roles/storage.admin":        {"user:martin.guibert@cloudskiff.com"},
						"roles/viewer":               {"serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com"},
						"roles/cloudasset.viewer":    {"serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com"},
						"roles/iam.securityReviewer": {"serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com"},
					},
				}, nil)
			},
			wantErr: nil,
		},
	}

	providerVersion := "3.78.0"
	resType := resource.ResourceType(googleresource.GoogleProjectIamMemberResourceType)
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

			realProvider, err := terraform2.InitTestGoogleProvider(providerLibrary, providerVersion)
			if err != nil {
				tt.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			managerRepository := &repository.MockCloudResourceManagerRepository{}
			if c.repositoryMock != nil {
				c.repositoryMock(managerRepository)
			}

			remoteLibrary.AddEnumerator(google.NewGoogleProjectIamMemberEnumerator(managerRepository, factory))

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
