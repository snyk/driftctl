package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

const BackendKeyCloud = "tfcloud"
const TerraformCloudAPI = "https://app.terraform.io/api/v2"

type Attributes struct {
	HostedStateDownloadUrl string `json:"hosted-state-download-url"`
}

type Data struct {
	Attributes Attributes `json:"attributes"`
}

type Body struct {
	Data Data `json:"data"`
}

func NewCloudReader(workspaceId string, opts *Options) (*HTTPBackend, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/workspaces/%s/current-state-version", TerraformCloudAPI, workspaceId), nil)

	req.Header.Add("Content-Type", "application/vnd.api+json")
	req.Header.Add("Authorization", opts.Headers["Authorization"])

	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	res, err := client.Do(req)

	if res.StatusCode == 404 {
		return nil, errors.Errorf("Error reading state from Terraform Cloud/Enterprise workspace: wrong workspace id")
	}

	if res.StatusCode == 401 {
		return nil, errors.Errorf("Error reading state from Terraform Cloud/Enterprise workspace: bad authentication token")
	}

	if err != nil {
		return nil, err
	}

	bodyBytes, _ := ioutil.ReadAll(res.Body)

	body := Body{}
	err = json.Unmarshal(bodyBytes, &body)

	if err != nil {
		fmt.Println("error:", err)
		panic(err.Error())
	}
	rawURL := body.Data.Attributes.HostedStateDownloadUrl

	opt := Options{}
	return NewHTTPReader(rawURL, &opt)
}
