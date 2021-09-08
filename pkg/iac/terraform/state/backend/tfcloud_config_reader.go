package backend

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mitchellh/go-homedir"
)

type container struct {
	Credentials struct {
		TerraformCloud struct {
			Token string
		} `json:"app.terraform.io"`
	}
}

type tfCloudConfigReader struct {
	reader io.ReadCloser
}

func NewTFCloudConfigReader(reader io.ReadCloser) *tfCloudConfigReader {
	return &tfCloudConfigReader{reader}
}

func (r *tfCloudConfigReader) GetToken() (string, error) {
	b, err := ioutil.ReadAll(r.reader)
	if err != nil {
		return "", errors.New("unable to read file")
	}

	var container container
	if err := json.Unmarshal(b, &container); err != nil {
		return "", err
	}
	if container.Credentials.TerraformCloud.Token == "" {
		return "", errors.New("driftctl could not read your Terraform configuration file, please check that this is a valid Terraform credentials file")
	}
	return container.Credentials.TerraformCloud.Token, nil
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
