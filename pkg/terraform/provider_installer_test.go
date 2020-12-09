package terraform

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/cloudskiff/driftctl/mocks"

	"github.com/stretchr/testify/assert"
)

func TestProviderInstallerGetAwsDoesNotExist(t *testing.T) {

	assert := assert.New(t)
	fakeTmpHome := t.TempDir()

	expectedSubFolder := "/.driftctl/plugins/linux_amd64"
	fakeUrl := "https://example.com"
	mockDownloader := mocks.ProviderDownloaderInterface{}
	mockDownloader.On("GetProviderUrl", "aws", "3.19.0").Return(fakeUrl)
	mockDownloader.On("Download", fakeUrl, path.Join(fakeTmpHome, expectedSubFolder)).Return(nil)

	installer := ProviderInstaller{
		downloader: &mockDownloader,
		homeDir:    fakeTmpHome,
	}

	providerPath, err := installer.GetAws()
	mockDownloader.AssertExpectations(t)

	assert.Nil(err)
	assert.Equal(path.Join(fakeTmpHome, expectedSubFolder, awsProviderName), providerPath)

}

func TestProviderInstallerGetAwsWithoutHomeDir(t *testing.T) {

	assert := assert.New(t)

	expectedHomeDir := os.TempDir()
	expectedSubFolder := "/.driftctl/plugins/linux_amd64"
	fakeUrl := "https://example.com"
	mockDownloader := mocks.ProviderDownloaderInterface{}
	mockDownloader.On("GetProviderUrl", "aws", "3.19.0").Return(fakeUrl)
	mockDownloader.On("Download", fakeUrl, path.Join(expectedHomeDir, expectedSubFolder)).Return(nil)

	installer := ProviderInstaller{
		downloader: &mockDownloader,
	}

	providerPath, err := installer.GetAws()
	mockDownloader.AssertExpectations(t)

	assert.Nil(err)
	assert.Equal(path.Join(expectedHomeDir, expectedSubFolder, awsProviderName), providerPath)

}

func TestProviderInstallerGetAwsAlreadyExist(t *testing.T) {

	assert := assert.New(t)
	fakeTmpHome := t.TempDir()
	expectedSubFolder := "/.driftctl/plugins/linux_amd64"
	err := os.MkdirAll(path.Join(fakeTmpHome, expectedSubFolder), 0755)
	if err != nil {
		t.Error(err)
	}
	_, err = os.Create(path.Join(fakeTmpHome, expectedSubFolder, awsProviderName))
	if err != nil {
		t.Error(err)
	}

	mockDownloader := mocks.ProviderDownloaderInterface{}

	installer := ProviderInstaller{
		downloader: &mockDownloader,
		homeDir:    fakeTmpHome,
	}

	providerPath, err := installer.GetAws()
	mockDownloader.AssertExpectations(t)

	assert.Nil(err)
	assert.Equal(path.Join(fakeTmpHome, expectedSubFolder, awsProviderName), providerPath)

}

func TestProviderInstallerGetAwsAlreadyExistButIsDirectory(t *testing.T) {

	assert := assert.New(t)
	fakeTmpHome := t.TempDir()
	expectedSubFolder := "/.driftctl/plugins/linux_amd64"
	invalidDirPath := path.Join(fakeTmpHome, expectedSubFolder, awsProviderName)
	err := os.MkdirAll(invalidDirPath, 0755)
	if err != nil {
		t.Error(err)
	}

	mockDownloader := mocks.ProviderDownloaderInterface{}

	installer := ProviderInstaller{
		downloader: &mockDownloader,
		homeDir:    fakeTmpHome,
	}

	providerPath, err := installer.GetAws()
	mockDownloader.AssertExpectations(t)

	assert.Empty(providerPath)
	assert.NotNil(err)
	assert.Equal(
		fmt.Sprintf(
			"found directory instead of provider binary in %s",
			invalidDirPath,
		),
		err.Error(),
	)

}
