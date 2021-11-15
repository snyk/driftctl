package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	pkghttp "github.com/cloudskiff/driftctl/pkg/http"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const BackendKeyTFCloud = "tfcloud"

type TFCloudAttributes struct {
	HostedStateDownloadUrl string `json:"hosted-state-download-url"`
}

type TFCloudData struct {
	Attributes TFCloudAttributes `json:"attributes"`
}

type TFCloudBody struct {
	Data TFCloudData `json:"data"`
}

type TFCloudBackend struct {
	request *http.Request
	client  pkghttp.HTTPClient
	reader  io.ReadCloser
	opts    *Options
}

func NewTFCloudReader(client pkghttp.HTTPClient, workspaceId string, opts *Options) (*TFCloudBackend, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/workspaces/%s/current-state-version", opts.TFCloudEndpoint, workspaceId), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/vnd.api+json")
	return &TFCloudBackend{req, client, nil, opts}, nil
}

func (t *TFCloudBackend) authorize() error {
	token := t.opts.TFCloudToken
	if token == "" {
		tfConfigFile, err := getTerraformConfigFile()
		if err != nil {
			return err
		}
		file, err := os.Open(tfConfigFile)
		if err != nil {
			return err
		}
		defer file.Close()
		reader := NewTFCloudConfigReader(file)
		token, err = reader.GetToken(t.request.URL.Host)
		if err != nil {
			return err
		}
	}
	t.request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	return nil
}

func (t *TFCloudBackend) Read(p []byte) (n int, err error) {
	if t.reader == nil {
		if err := t.authorize(); err != nil {
			return 0, err
		}
		res, err := t.client.Do(t.request)
		if err != nil {
			return 0, err
		}

		if res.StatusCode < 200 || res.StatusCode >= 400 {
			return 0, errors.Errorf("error requesting terraform cloud backend state: status code: %d", res.StatusCode)
		}

		body := TFCloudBody{}
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		err = json.Unmarshal(bodyBytes, &body)
		if err != nil {
			return 0, err
		}

		rawURL := body.Data.Attributes.HostedStateDownloadUrl
		logrus.WithFields(logrus.Fields{"hosted-state-download-url": rawURL}).Trace("Terraform Cloud backend response")

		h, err := NewHTTPReader(t.client, rawURL, &Options{})
		if err != nil {
			return 0, err
		}
		t.reader = h
	}
	return t.reader.Read(p)
}

func (t *TFCloudBackend) Close() error {
	if t.reader != nil {
		return t.reader.Close()
	}
	return errors.New("Unable to close reader as nothing was opened")
}
