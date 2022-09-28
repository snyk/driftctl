package backend

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mitchellh/go-homedir"
)

type container struct {
	Credentials map[string]containerToken
}

type containerToken struct {
	Token string
}

type tfCloudConfigReader struct {
	reader io.ReadCloser
}

func NewTFCloudConfigReader(reader io.ReadCloser) *tfCloudConfigReader {
	return &tfCloudConfigReader{reader}
}

func (r *tfCloudConfigReader) GetToken(host string) (string, error) {
	b, err := io.ReadAll(r.reader)
	if err != nil {
		return "", errors.New("unable to read file")
	}

	var container container
	if err := json.Unmarshal(b, &container); err != nil {
		return "", err
	}
	if container.Credentials[host].Token == "" {
		return "", errors.New("driftctl could not read your Terraform configuration file, please check that this is a valid Terraform credentials file")
	}
	return container.Credentials[host].Token, nil
}

func getTerraformConfigFile() (string, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	basePath := filepath.Join(homeDir, ".terraform.d")
	if runtime.GOOS == "windows" {
		basePath = filepath.Join(os.Getenv("APPDATA"), "terraform.d")
	}
	return filepath.Join(basePath, "credentials.tfrc.json"), nil
}
