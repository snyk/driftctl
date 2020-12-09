package terraform

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
)

const awsProviderName = "terraform-provider-aws_v3.19.0_x5"

type HomeDirInterface interface {
	Dir() (string, error)
}

type ProviderInstaller struct {
	downloader ProviderDownloaderInterface
	homeDir    string
}

func NewProviderInstaller() (*ProviderInstaller, error) {
	homedir, err := homedir.Dir()
	if err != nil {
		homedir = ""
	}
	return &ProviderInstaller{
		NewProviderDownloader(),
		homedir,
	}, nil
}

func (p *ProviderInstaller) GetAws() (string, error) {
	if p.homeDir == "" {
		p.homeDir = os.TempDir()
	}
	providerDir := path.Join(p.homeDir, fmt.Sprintf("/.driftctl/plugins/%s_%s/", runtime.GOOS, runtime.GOARCH))
	providerPath := path.Join(providerDir, awsProviderName)

	info, err := os.Stat(providerPath)
	if err != nil && os.IsNotExist(err) {
		logrus.WithFields(logrus.Fields{
			"path": providerPath,
		}).Debug("AWS provider not found, downloading ...")
		fmt.Printf("Downloading AWS terraform provider: %s\n", awsProviderName)
		err := p.downloader.Download(
			p.downloader.GetProviderUrl(AWS, "3.19.0"),
			providerDir,
		)
		if err != nil {
			return "", err
		}
		logrus.Debug("Download successful")
	}

	if info != nil && info.IsDir() {
		return "", fmt.Errorf("found directory instead of provider binary in %s", providerPath)
	}

	if info != nil {
		logrus.WithFields(logrus.Fields{
			"path": providerPath,
		}).Debug("Found existing aws provider")
	}

	return providerPath, nil
}
