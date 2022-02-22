package azure

import (
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/pkg/errors"
)

func GetBlobSharedKey() (*azblob.SharedKeyCredential, error) {
	storageAccountName, exist := os.LookupEnv("AZURE_STORAGE_ACCOUNT")
	if !exist {
		return nil, errors.New("AZURE_STORAGE_ACCOUNT should be defined to be able to read state from azure backend")
	}

	storageAccountKey, exist := os.LookupEnv("AZURE_STORAGE_KEY")
	if !exist {
		return nil, errors.New("AZURE_STORAGE_KEY should be defined to be able to read state from azure backend")
	}

	credential, err := azblob.NewSharedKeyCredential(storageAccountName, storageAccountKey)
	if err != nil {
		return nil, err
	}

	return credential, nil
}
