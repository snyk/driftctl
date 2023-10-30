package remote

import (
	"context"
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/remote/google"
	"github.com/snyk/driftctl/enumeration/remote/google/repository"
	googleresource "github.com/snyk/driftctl/enumeration/resource/google"
	"github.com/snyk/driftctl/enumeration/terraform"

	asset "cloud.google.com/go/asset/apiv1"
	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/mocks"

	assetpb "cloud.google.com/go/asset/apiv1/assetpb"
	"github.com/snyk/driftctl/test/goldenfile"
	testgoogle "github.com/snyk/driftctl/test/google"
	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGoogleStorageBucket(t *testing.T) {
	cases := []struct {
		test             string
		dirName          string
		response         []*assetpb.ResourceSearchResult
		responseErr      error
		setupAlerterMock func(alerter *mocks.AlerterInterface)
		assertExpected   func(*testing.T, []*resource.Resource)
		wantErr          error
	}{
		{
			test:     "no storage buckets",
			dirName:  "google_storage_bucket_empty",
			response: []*assetpb.ResourceSearchResult{},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: nil,
		},
		{
			test:    "multiples storage buckets",
			dirName: "google_storage_bucket",
			response: []*assetpb.ResourceSearchResult{
				{
					AssetType:   "storage.googleapis.com/Bucket",
					DisplayName: "driftctl-unittest-1",
				},
				{
					AssetType:   "storage.googleapis.com/Bucket",
					DisplayName: "driftctl-unittest-2",
				},
				{
					AssetType:   "storage.googleapis.com/Bucket",
					DisplayName: "driftctl-unittest-3",
				},
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 3)

				assert.Equal(t, "driftctl-unittest-1", got[0].ResourceId())
				assert.Equal(t, googleresource.GoogleStorageBucketResourceType, got[0].ResourceType())

				assert.Equal(t, "driftctl-unittest-2", got[1].ResourceId())
				assert.Equal(t, googleresource.GoogleStorageBucketResourceType, got[1].ResourceType())

				assert.Equal(t, "driftctl-unittest-3", got[2].ResourceId())
				assert.Equal(t, googleresource.GoogleStorageBucketResourceType, got[2].ResourceType())
			},
			wantErr: nil,
		},
		{
			test:        "cannot list storage buckets",
			dirName:     "google_storage_bucket_empty",
			responseErr: status.Error(codes.PermissionDenied, "The caller does not have permission"),
			setupAlerterMock: func(alerter *mocks.AlerterInterface) {
				alerter.On(
					"SendAlert",
					"google_storage_bucket",
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							status.Error(codes.PermissionDenied, "The caller does not have permission"),
							"google_storage_bucket",
						),
						alerts.EnumerationPhase,
					),
				).Once()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: nil,
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			if c.setupAlerterMock != nil {
				c.setupAlerterMock(alerter)
			}

			var assetClient *asset.Client
			if !shouldUpdate {
				var err error
				assetClient, err = testgoogle.NewFakeAssetServer(c.response, c.responseErr)
				if err != nil {
					tt.Fatal(err)
				}
			}

			realProvider, err := terraform2.InitTestGoogleProvider(providerLibrary, "3.78.0")
			if err != nil {
				tt.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				ctx := context.Background()
				assetClient, err = asset.NewClient(ctx)
				if err != nil {
					tt.Fatal(err)
				}
				err = realProvider.Init()
				if err != nil {
					tt.Fatal(err)
				}
				provider.ShouldUpdate()
			}

			repo := repository.NewAssetRepository(assetClient, realProvider.GetConfig(), cache.New(0))

			remoteLibrary.AddEnumerator(google.NewGoogleStorageBucketEnumerator(repo, factory))

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
			c.assertExpected(tt, got)
		})
	}
}

func TestGoogleStorageBucketIAMMember(t *testing.T) {
	cases := []struct {
		test                  string
		dirName               string
		assetRepositoryMock   func(assetRepository *repository.MockAssetRepository)
		storageRepositoryMock func(storageRepository *repository.MockStorageRepository)
		responseErr           error
		setupAlerterMock      func(alerter *mocks.AlerterInterface)
		assertExpected        func(*testing.T, []*resource.Resource)
		wantErr               error
	}{
		{
			test:    "no storage buckets",
			dirName: "google_storage_bucket_member_empty",
			assetRepositoryMock: func(assetRepository *repository.MockAssetRepository) {
				assetRepository.On("SearchAllBuckets").Return([]*assetpb.ResourceSearchResult{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: nil,
		},
		{
			test:    "multiples storage buckets, no bindings",
			dirName: "google_storage_bucket_member_empty",
			assetRepositoryMock: func(assetRepository *repository.MockAssetRepository) {
				assetRepository.On("SearchAllBuckets").Return([]*assetpb.ResourceSearchResult{
					{
						AssetType:   "storage.googleapis.com/Bucket",
						DisplayName: "dctlgstoragebucketiambinding-1",
					},
					{
						AssetType:   "storage.googleapis.com/Bucket",
						DisplayName: "dctlgstoragebucketiambinding-2",
					},
				}, nil)
			},
			storageRepositoryMock: func(storageRepository *repository.MockStorageRepository) {
				storageRepository.On("ListAllBindings", "dctlgstoragebucketiambinding-1").Return(map[string][]string{}, nil)
				storageRepository.On("ListAllBindings", "dctlgstoragebucketiambinding-2").Return(map[string][]string{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: nil,
		},
		{
			test:    "Cannot list bindings",
			dirName: "google_storage_bucket_member_listing_error",
			assetRepositoryMock: func(assetRepository *repository.MockAssetRepository) {
				assetRepository.On("SearchAllBuckets").Return([]*assetpb.ResourceSearchResult{
					{
						AssetType:   "storage.googleapis.com/Bucket",
						DisplayName: "dctlgstoragebucketiambinding-1",
					},
				}, nil)
			},
			storageRepositoryMock: func(storageRepository *repository.MockStorageRepository) {
				storageRepository.On("ListAllBindings", "dctlgstoragebucketiambinding-1").Return(
					map[string][]string{},
					errors.New("googleapi: Error 403: driftctl-acc-circle@driftctl-qa-1.iam.gserviceaccount.com does not have storage.buckets.getIamPolicy access to the Google Cloud Storage bucket., forbidden"))
			},
			setupAlerterMock: func(alerter *mocks.AlerterInterface) {
				alerter.On(
					"SendAlert",
					"google_storage_bucket_iam_member",
					alerts.NewRemoteAccessDeniedAlert(
						common.RemoteGoogleTerraform,
						remoteerr.NewResourceListingError(
							errors.New("googleapi: Error 403: driftctl-acc-circle@driftctl-qa-1.iam.gserviceaccount.com does not have storage.buckets.getIamPolicy access to the Google Cloud Storage bucket., forbidden"),
							"google_storage_bucket_iam_member",
						),
						alerts.EnumerationPhase,
					),
				).Once()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: nil,
		},
		{
			test:    "multiples storage buckets, multiple bindings",
			dirName: "google_storage_bucket_member_listing_multiple",
			assetRepositoryMock: func(assetRepository *repository.MockAssetRepository) {
				assetRepository.On("SearchAllBuckets").Return([]*assetpb.ResourceSearchResult{
					{
						AssetType:   "storage.googleapis.com/Bucket",
						DisplayName: "dctlgstoragebucketiambinding-1",
					},
					{
						AssetType:   "storage.googleapis.com/Bucket",
						DisplayName: "dctlgstoragebucketiambinding-2",
					},
				}, nil)
			},
			storageRepositoryMock: func(storageRepository *repository.MockStorageRepository) {
				storageRepository.On("ListAllBindings", "dctlgstoragebucketiambinding-1").Return(map[string][]string{
					"roles/storage.admin":        {"user:elie.charra@cloudskiff.com"},
					"roles/storage.objectViewer": {"user:william.beuil@cloudskiff.com"},
				}, nil)

				storageRepository.On("ListAllBindings", "dctlgstoragebucketiambinding-2").Return(map[string][]string{
					"roles/storage.admin":        {"user:william.beuil@cloudskiff.com"},
					"roles/storage.objectViewer": {"user:elie.charra@cloudskiff.com"},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 4)

				var resourceIds []string
				for _, res := range got {
					assert.Equal(t, googleresource.GoogleStorageBucketIamMemberResourceType, res.ResourceType())
					resourceIds = append(resourceIds, res.ResourceId())
				}

				assert.Contains(t, resourceIds, "b/dctlgstoragebucketiambinding-1/roles/storage.admin/user:elie.charra@cloudskiff.com")
				assert.Contains(t, resourceIds, "b/dctlgstoragebucketiambinding-1/roles/storage.objectViewer/user:william.beuil@cloudskiff.com")
				assert.Contains(t, resourceIds, "b/dctlgstoragebucketiambinding-2/roles/storage.admin/user:william.beuil@cloudskiff.com")
				assert.Contains(t, resourceIds, "b/dctlgstoragebucketiambinding-2/roles/storage.admin/user:william.beuil@cloudskiff.com")
			},
			wantErr: nil,
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			repositoryCache := cache.New(100)

			shouldUpdate := c.dirName == *goldenfile.Update

			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			if c.setupAlerterMock != nil {
				c.setupAlerterMock(alerter)
			}

			storageRepo := &repository.MockStorageRepository{}
			if c.storageRepositoryMock != nil {
				c.storageRepositoryMock(storageRepo)
			}
			var storageRepository repository.StorageRepository = storageRepo
			if shouldUpdate {
				storageClient, err := storage.NewClient(context.Background())
				if err != nil {
					panic(err)
				}
				storageRepository = repository.NewStorageRepository(storageClient, repositoryCache)
			}

			assetRepo := &repository.MockAssetRepository{}
			if c.assetRepositoryMock != nil {
				c.assetRepositoryMock(assetRepo)
			}
			var assetRepository repository.AssetRepository = assetRepo

			realProvider, err := terraform2.InitTestGoogleProvider(providerLibrary, "3.78.0")
			if err != nil {
				tt.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			remoteLibrary.AddEnumerator(google.NewGoogleStorageBucketIamMemberEnumerator(assetRepository, storageRepository, factory))

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
			c.assertExpected(tt, got)
		})
	}
}
