package remote

import (
	"errors"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/snyk/driftctl/mocks"
	"github.com/snyk/driftctl/pkg/filter"
	"github.com/snyk/driftctl/pkg/remote/alerts"
	"github.com/snyk/driftctl/pkg/remote/aws"
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	"github.com/snyk/driftctl/pkg/remote/common"
	remoteerr "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	resourceaws "github.com/snyk/driftctl/pkg/resource/aws"
	"github.com/snyk/driftctl/pkg/terraform"
	testresource "github.com/snyk/driftctl/test/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestElastiCacheCluster(t *testing.T) {
	dummyError := errors.New("dummy error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockElastiCacheRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no elasticache clusters",
			mocks: func(repository *repository.MockElastiCacheRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllCacheClusters").Return([]*elasticache.CacheCluster{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "should list elasticache clusters",
			mocks: func(repository *repository.MockElastiCacheRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllCacheClusters").Return([]*elasticache.CacheCluster{
					{CacheClusterId: awssdk.String("cluster-foo")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)
				assert.Equal(t, got[0].ResourceId(), "cluster-foo")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsElastiCacheClusterResourceType)
			},
		},
		{
			test: "cannot list elasticache clusters (403)",
			mocks: func(repository *repository.MockElastiCacheRepository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllCacheClusters").Return(nil, awsError)
				alerter.On("SendAlert", resourceaws.AwsElastiCacheClusterResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsElastiCacheClusterResourceType, resourceaws.AwsElastiCacheClusterResourceType), alerts.EnumerationPhase)).Return()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "cannot list elasticache clusters (dummy error)",
			mocks: func(repository *repository.MockElastiCacheRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllCacheClusters").Return(nil, dummyError)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: remoteerr.NewResourceScanningError(dummyError, resourceaws.AwsElastiCacheClusterResourceType, ""),
		},
	}

	providerVersion := "3.19.0"
	schemaRepository := testresource.InitFakeSchemaRepository("aws", providerVersion)
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			scanOptions := ScannerOptions{}
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockElastiCacheRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ElastiCacheRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewElastiCacheClusterEnumerator(repo, factory))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)
			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}
