package terraform

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

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
	providerDir := p.getProviderDirectory()
	providerPath := p.getBinaryPath()

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

	return p.getBinaryPath(), nil
}

func (p ProviderInstaller) getProviderDirectory() string {
	if p.homeDir == "" {
		p.homeDir = os.TempDir()
	}
	return path.Join(p.homeDir, fmt.Sprintf("/.driftctl/plugins/%s_%s/", runtime.GOOS, runtime.GOARCH))
}

// Handle postfixes in binary names
func (p *ProviderInstaller) getBinaryPath() string {
	providerDir := p.getProviderDirectory()
	binaryName := p.config.GetBinaryName()
	_, err := os.Stat(path.Join(providerDir, binaryName))
	if err != nil && os.IsNotExist(err) {
		_ = filepath.WalkDir(providerDir, func(filePath string, d fs.DirEntry, err error) error {
			if d != nil && strings.HasPrefix(d.Name(), p.config.GetBinaryName()) {
				binaryName = d.Name()
			}
			return nil
		})
	}

	return path.Join(providerDir, binaryName)
}
