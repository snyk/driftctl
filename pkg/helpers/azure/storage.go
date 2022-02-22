package azure

import (
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

func NewStorageAccountsClient() (*armstorage.StorageAccountsClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	return armstorage.NewStorageAccountsClient(os.Getenv("AZURE_SUBSCRIPTION_ID"), cred, nil), nil

}
