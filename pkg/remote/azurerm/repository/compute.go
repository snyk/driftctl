package repository

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/snyk/driftctl/pkg/remote/azurerm/common"
	"github.com/snyk/driftctl/pkg/remote/cache"
)

type ComputeRepository interface {
	ListAllImages() ([]*armcompute.Image, error)
	ListAllSSHPublicKeys() ([]*armcompute.SSHPublicKeyResource, error)
}

type imagesListPager interface {
	pager
	PageResponse() armcompute.ImagesListResponse
}

type imagesClient interface {
	List(options *armcompute.ImagesListOptions) imagesListPager
}

type imagesClientImpl struct {
	client *armcompute.ImagesClient
}

func (c imagesClientImpl) List(options *armcompute.ImagesListOptions) imagesListPager {
	return c.client.List(options)
}

type sshPublicKeyListPager interface {
	pager
	PageResponse() armcompute.SSHPublicKeysListBySubscriptionResponse
}

type sshPublicKeyClient interface {
	ListBySubscription(options *armcompute.SSHPublicKeysListBySubscriptionOptions) sshPublicKeyListPager
}

type sshPublicKeyClientImpl struct {
	client *armcompute.SSHPublicKeysClient
}

func (c sshPublicKeyClientImpl) ListBySubscription(options *armcompute.SSHPublicKeysListBySubscriptionOptions) sshPublicKeyListPager {
	return c.client.ListBySubscription(options)
}

type computeRepository struct {
	imagesClient       imagesClient
	sshPublicKeyClient sshPublicKeyClient
	cache              cache.Cache
}

func NewComputeRepository(cred azcore.TokenCredential, options *arm.ClientOptions, config common.AzureProviderConfig, cache cache.Cache) *computeRepository {
	return &computeRepository{
		&imagesClientImpl{armcompute.NewImagesClient(config.SubscriptionID, cred, options)},
		&sshPublicKeyClientImpl{armcompute.NewSSHPublicKeysClient(config.SubscriptionID, cred, options)},
		cache,
	}
}

func (s *computeRepository) ListAllImages() ([]*armcompute.Image, error) {
	cacheKey := "computeListAllImages"
	if v := s.cache.Get(cacheKey); v != nil {
		return v.([]*armcompute.Image), nil
	}

	pager := s.imagesClient.List(nil)
	results := make([]*armcompute.Image, 0)
	for pager.NextPage(context.Background()) {
		resp := pager.PageResponse()
		if err := pager.Err(); err != nil {
			return nil, err
		}
		results = append(results, resp.Value...)
	}
	if err := pager.Err(); err != nil {
		return nil, err
	}

	s.cache.Put(cacheKey, results)
	return results, nil
}

func (s *computeRepository) ListAllSSHPublicKeys() ([]*armcompute.SSHPublicKeyResource, error) {
	cacheKey := "computeListAllSSHPublicKeys"
	if v := s.cache.Get(cacheKey); v != nil {
		return v.([]*armcompute.SSHPublicKeyResource), nil
	}

	pager := s.sshPublicKeyClient.ListBySubscription(nil)
	results := make([]*armcompute.SSHPublicKeyResource, 0)
	for pager.NextPage(context.Background()) {
		resp := pager.PageResponse()
		if err := pager.Err(); err != nil {
			return nil, err
		}
		results = append(results, resp.Value...)
	}
	if err := pager.Err(); err != nil {
		return nil, err
	}

	s.cache.Put(cacheKey, results)
	return results, nil
}
