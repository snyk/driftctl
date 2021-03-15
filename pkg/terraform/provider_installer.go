package terraform

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/output"
)

type HomeDirInterface interface {
	Dir() (string, error)
}

type ProviderInstaller struct {
	downloader ProviderDownloaderInterface
	config     ProviderConfig
	homeDir    string
}

func NewProviderInstaller(config ProviderConfig) (*ProviderInstaller, error) {
	homedir, err := homedir.Dir()
	if err != nil {
		homedir = ""
	}
	return &ProviderInstaller{
		NewProviderDownloader(),
		config,
		homedir,
	}, nil
}

func (p *ProviderInstaller) Install() (string, error) {
	if p.homeDir == "" {
		p.homeDir = os.TempDir()
	}
	providerDir := path.Join(p.homeDir, fmt.Sprintf("/.driftctl/plugins/%s_%s/", runtime.GOOS, runtime.GOARCH))
	providerPath := path.Join(providerDir, p.config.GetBinaryName())

	info, err := os.Stat(providerPath)
	if err != nil && os.IsNotExist(err) {
		logrus.WithFields(logrus.Fields{
			"path": providerPath,
		}).Debug("provider not found, downloading ...")
		output.Printf("Downloading terraform provider: %s\n", p.config.Key)
		err := p.downloader.Download(
			p.config.GetDownloadUrl(),
			providerDir,
		)
		if err != nil {
			return "", err
		}
		logrus.Debug("Download successful")
	}

	if info != nil && info.IsDir() {
		return "", errors.Errorf(
			"found directory instead of provider binary in %s",
			providerPath,
		)
	}

	if info != nil {
		logrus.WithFields(logrus.Fields{
			"path": providerPath,
		}).Debug("Found existing provider")
	}

	return providerPath, nil
}
