package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	pkghttp "github.com/cloudskiff/driftctl/pkg/http"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const BackendKeyTFCloud = "tfcloud"
const TFCloudAPI = "https://app.terraform.io/api/v2"

type TFCloudAttributes struct {
	HostedStateDownloadUrl string `json:"hosted-state-download-url"`
}

type TFCloudData struct {
	Attributes TFCloudAttributes `json:"attributes"`
}

type TFCloudBody struct {
	Data TFCloudData `json:"data"`
}

func NewTFCloudReader(client pkghttp.HTTPClient, workspaceId string, opts *Options) (*HTTPBackend, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/workspaces/%s/current-state-version", TFCloudAPI, workspaceId), nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/vnd.api+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", opts.TFCloudToken))

	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return nil, errors.Errorf("error requesting terraform cloud backend state: status code: %d", res.StatusCode)
	}

	bodyBytes, _ := ioutil.ReadAll(res.Body)

	body := TFCloudBody{}
	err = json.Unmarshal(bodyBytes, &body)

	if err != nil {
		return nil, err
	}

	rawURL := body.Data.Attributes.HostedStateDownloadUrl
	logrus.WithFields(logrus.Fields{"hosted-state-download-url": rawURL}).Trace("Terraform Cloud backend response")

	opt := Options{}
	return NewHTTPReader(client, rawURL, &opt)
}
