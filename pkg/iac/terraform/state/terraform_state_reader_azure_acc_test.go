package state_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/pkg/helpers/azure"
	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/acceptance"
)

func TestAcc_StateReader_WithMultiplesStatesInAzure(t *testing.T) {
	// WARNING: If you change the resource group you also have to change it in terraform files
	resourceGroupName := "driftctl-qa-1"
	storageAccount := "driftctlacctest"
	containerName := "foobar"
	checkEnv := map[string]string{
		"AZURE_STORAGE_ACCOUNT": storageAccount,
	}
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		OnStart: func() {
			// Remove existing storage account if it already exists
			_ = removeAzureStorageAccount(resourceGroupName, storageAccount)
			key, err := createAzureStorageContainer(resourceGroupName, storageAccount, containerName)
			if err != nil {
				t.Fatal(err)
			}
			checkEnv["AZURE_STORAGE_KEY"] = key
		},
		Paths: []string{"./testdata/acc/multiples_states_azure/container_registry", "./testdata/acc/multiples_states_azure/another_container_registry"},
		Args: []string{
			"scan",
			"--from", fmt.Sprintf("tfstate+azurerm://%s/states/valid/**", containerName),
			"--to", "azure+tf",
			"--filter", "Type=='azurerm_container_registry'",
		},
		Checks: []acceptance.AccCheck{
			{
				Env: checkEnv,
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertInfrastructureIsInSync()
					result.Equal(2, result.Summary().TotalManaged)
					result.Equal("azurerm_container_registry", result.Managed()[0].ResourceType())
					result.Equal("another_registry", result.Managed()[0].Source.InternalName())
					result.Equal("tfstate+azurerm://foobar/states/valid/another_container_registry/terraform.tfstate", result.Managed()[0].Source.Source())
					result.Equal("azurerm_container_registry", result.Managed()[1].ResourceType())
					result.Equal("registry", result.Managed()[1].Source.InternalName())
					result.Equal("tfstate+azurerm://foobar/states/valid/registry/terraform.tfstate", result.Managed()[1].Source.Source())
				},
			},
		},
		OnEnd: func() {
			err := removeAzureStorageAccount(resourceGroupName, storageAccount)
			if err != nil {
				t.Fatal(err)
			}
		},
	})
}

func createAzureStorageContainer(resourceGroupName, storageAccount, containerName string) (string, error) {
	// Let's begin by creating a new storage account
	client, err := azure.NewStorageAccountsClient()
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	poller, err := client.BeginCreate(
		ctx,
		resourceGroupName,
		storageAccount,
		armstorage.StorageAccountCreateParameters{
			SKU: &armstorage.SKU{
				Name: func() *armstorage.SKUName { sku := armstorage.SKUNameStandardLRS; return &sku }(),
			},
			Kind:     func() *armstorage.Kind { kind := armstorage.KindStorageV2; return &kind }(),
			Location: to.StringPtr("westeurope"),
		},
		nil,
	)
	if err != nil {
		return "", err
	}

	_, err = poller.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return "", err
	}

	// Retrieve key from storage account
	keys, err := client.ListKeys(ctx, resourceGroupName, storageAccount, nil)
	if err != nil {
		return "", err
	}
	if len(keys.Keys) == 0 {
		return "", errors.Errorf("Unable to retrieve keys for storage account %s", storageAccount)
	}
	key := *keys.Keys[0].Value

	// Create a blob container
	cred, err := azblob.NewSharedKeyCredential(storageAccount, key)
	if err != nil {
		return "", err
	}
	blobClient, err := azblob.NewServiceClientWithSharedKey(fmt.Sprintf("https://%s.blob.core.windows.net/", storageAccount), cred, nil)
	if err != nil {
		return "", err
	}
	_, err = blobClient.CreateContainer(ctx, containerName, nil)
	if err != nil {
		return "", err
	}

	return key, nil
}

func removeAzureStorageAccount(resourceGroupName, storageAccount string) error {
	client, err := azure.NewStorageAccountsClient()
	if err != nil {
		return err
	}
	_, err = client.Delete(context.Background(), resourceGroupName, storageAccount, nil)
	return err
}
