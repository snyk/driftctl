package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/armcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/armstorage"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm/common"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type StorageRespository interface {
	ListAllStorageAccount() ([]*armstorage.StorageAccount, error)
	ListAllStorageContainer(account *armstorage.StorageAccount) ([]string, error)
}

// Interfaces are only used to create mock on Azure SDK
type blobContainerClient interface {
	List(resourceGroupName string, accountName string, options *armstorage.BlobContainersListOptions) armstorage.ListContainerItemsPager
}

type storageAccountClient interface {
	List(options *armstorage.StorageAccountsListOptions) armstorage.StorageAccountListResultPager
}

type storageRepository struct {
	listAllStorageAccountLock sync.Locker
	storageAccountsClient     storageAccountClient
	blobContainerClient       blobContainerClient
	cache                     cache.Cache
}

func NewStorageRepository(con *armcore.Connection, config common.AzureProviderConfig, cache cache.Cache) *storageRepository {
	return &storageRepository{
		&sync.Mutex{},
		armstorage.NewStorageAccountsClient(con, config.SubscriptionID),
		armstorage.NewBlobContainersClient(con, config.SubscriptionID),
		cache,
	}
}

func (s *storageRepository) ListAllStorageAccount() ([]*armstorage.StorageAccount, error) {

	// Since ListAllStorageAccount can be called from multiple suppliers we should lock here to ensure
	// the cache is hit when multiple calls are running in parallel
	s.listAllStorageAccountLock.Lock()
	defer s.listAllStorageAccountLock.Unlock()

	if v := s.cache.Get("ListAllStorageAccount"); v != nil {
		return v.([]*armstorage.StorageAccount), nil
	}

	pager := s.storageAccountsClient.List(nil)
	results := make([]*armstorage.StorageAccount, 0)
	for pager.NextPage(context.Background()) {
		resp := pager.PageResponse()
		if err := pager.Err(); err != nil {
			return nil, err
		}
		results = append(results, resp.StorageAccountListResult.Value...)
	}

	if err := pager.Err(); err != nil {
		return nil, err
	}

	s.cache.Put("ListAllStorageAccount", results)

	return results, nil
}

func (s *storageRepository) ListAllStorageContainer(account *armstorage.StorageAccount) ([]string, error) {

	cacheKey := fmt.Sprintf("ListAllStorageContainer_%s", *account.Name)
	if v := s.cache.Get(cacheKey); v != nil {
		return v.([]string), nil
	}

	res, err := azure.ParseResourceID(*account.ID)
	if err != nil {
		return nil, err
	}

	pager := s.blobContainerClient.List(res.ResourceGroup, *account.Name, nil)
	results := make([]string, 0)
	for pager.NextPage(context.Background()) {
		resp := pager.PageResponse()
		if err := pager.Err(); err != nil {
			return nil, err
		}
		for _, item := range resp.ListContainerItems.Value {
			results = append(results, fmt.Sprintf("%s%s", *account.Properties.PrimaryEndpoints.Blob, *item.Name))
		}
	}

	if err := pager.Err(); err != nil {
		return nil, err
	}

	s.cache.Put(cacheKey, results)

	return results, nil
}
