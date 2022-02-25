package backend

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/pkg/iac/terraform/state/backend/options"
)

const BackendKeyAzureRM = "azurerm"

type AzureRMBackend struct {
	reader        io.ReadCloser
	storageClient azblob.BlockBlobClient
}

func NewAzureRMReader(path string, opts options.AzureRMBackendOptions) (*AzureRMBackend, error) {
	bucketPath := strings.Split(path, "/")
	if len(bucketPath) < 2 || bucketPath[1] == "" {
		return nil, errors.Errorf("Unable to parse azurerm backend storage path: %s. Must be CONTAINER/PATH/TO/OBJECT", path)
	}
	containerName := bucketPath[0]
	objectPath := strings.Join(bucketPath[1:], "/")

	credential, err := azblob.NewSharedKeyCredential(opts.StorageAccount, opts.StorageKey)
	if err != nil {
		return nil, err
	}

	blobClient, err := azblob.NewBlockBlobClientWithSharedKey(
		fmt.Sprintf(
			"https://%s.blob.core.windows.net/%s/%s",
			credential.AccountName(),
			containerName,
			objectPath,
		),
		credential,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &AzureRMBackend{
		storageClient: blobClient,
	}, nil
}

func (s *AzureRMBackend) Read(p []byte) (int, error) {
	if s.reader == nil {
		ctx := context.Background()
		data, err := s.storageClient.Download(ctx, nil)
		if err != nil {
			return 0, err
		}
		s.reader = data.Body(azblob.RetryReaderOptions{})
	}
	return s.reader.Read(p)
}

func (s *AzureRMBackend) Close() error {
	if s.reader != nil {
		return s.reader.Close()
	}
	return errors.New("Unable to close reader as nothing was opened")
}
