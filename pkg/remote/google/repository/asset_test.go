package repository

import (
	"testing"

	"github.com/snyk/driftctl/pkg/remote/cache"
	"github.com/snyk/driftctl/pkg/remote/google/config"
	"github.com/snyk/driftctl/test/google"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	assetpb "google.golang.org/genproto/googleapis/cloud/asset/v1"
)

func Test_assetRepository_searchAllResources_CacheHit(t *testing.T) {

	expectedResults := []*assetpb.ResourceSearchResult{
		{
			AssetType:   "google_fake_type",
			DisplayName: "driftctl-unittest-1",
		},
		{
			AssetType:   "google_another_fake_type",
			DisplayName: "driftctl-unittest-1",
		},
	}

	c := &cache.MockCache{}
	c.On("GetAndLock", "SearchAllResources_").Return(expectedResults).Times(1)
	c.On("Unlock", "SearchAllResources_").Times(1)
	repo := NewAssetRepository(nil, config.GCPTerraformConfig{Scopes: []string{""}}, c)

	got, err := repo.searchAllResources("google_fake_type")
	c.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Len(t, got, 1)
}

func Test_assetRepository_searchAllResources_CacheMiss(t *testing.T) {

	expectedResults := []*assetpb.ResourceSearchResult{
		{
			AssetType:   "google_fake_type",
			DisplayName: "driftctl-unittest-1",
		},
		{
			AssetType:   "google_another_fake_type",
			DisplayName: "driftctl-unittest-1",
		},
	}
	assetClient, err := google.NewFakeAssetServer(expectedResults, nil)
	if err != nil {
		t.Fatal(err)
	}
	c := &cache.MockCache{}
	c.On("GetAndLock", "SearchAllResources_").Return(nil).Times(1)
	c.On("Unlock", "SearchAllResources_").Times(1)
	c.On("Put", "SearchAllResources_", mock.IsType([]*assetpb.ResourceSearchResult{})).Return(false).Times(1)
	repo := NewAssetRepository(assetClient, config.GCPTerraformConfig{Scopes: []string{""}}, c)

	got, err := repo.searchAllResources("google_fake_type")
	c.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Len(t, got, 1)
}
