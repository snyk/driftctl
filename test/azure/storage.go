package azure

import "github.com/Azure/azure-sdk-for-go/sdk/storage/armstorage"

type StorageAccountPager interface {
	armstorage.StorageAccountListResultPager
}

type ListContainerItemPager interface {
	armstorage.ListContainerItemsPager
}
