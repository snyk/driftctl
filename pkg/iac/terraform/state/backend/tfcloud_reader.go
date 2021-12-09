package backend

import (
	"bytes"
	"context"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/pkg/errors"
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
	client        *tfe.Client
	reader        io.ReadCloser
	opts          *Options
	workspacePath string
}

func NewTFCloudReader(workspacePath string, opts *Options) *TFCloudBackend {
	return &TFCloudBackend{opts: opts, workspacePath: workspacePath}
}

func (t *TFCloudBackend) getToken() (string, error) {
	token := t.opts.TFCloudToken
	if token == "" {
		tfConfigFile, err := getTerraformConfigFile()
		if err != nil {
			return "", err
		}

		file, err := os.Open(tfConfigFile)
		if err != nil {
			return "", err
		}
		defer file.Close()
		reader := NewTFCloudConfigReader(file)

		u, err := url.Parse(t.opts.TFCloudEndpoint)
		if err != nil {
			return "", err
		}
		return reader.GetToken(u.Host)
	}
	return token, nil
}

// A regular expression used to validate string workspace ID patterns.
var reStringID = regexp.MustCompile(`^ws-[a-zA-Z0-9\-\._]+$`)

// isValidWorkspaceID checks if the given input is present and non-empty.
func isValidWorkspaceID(v string) bool {
	return v != "" && reStringID.MatchString(v)
}

func (t *TFCloudBackend) getWorkspaceId() (string, error) {
	if isValidWorkspaceID(t.workspacePath) {
		return t.workspacePath, nil
	}
	workspacePath := strings.Split(t.workspacePath, "/")
	if len(workspacePath) != 2 {
		return "", errors.New("unable to parse terraform cloud workspace, it should be either a workspace id (ws-xxxxx) or a {org}/{workspaceName}")
	}
	workspace, err := t.client.Workspaces.Read(context.Background(), workspacePath[0], workspacePath[1])
	if err != nil {
		return "", errors.Errorf("unable to read terraform workspace id: %s", err.Error())
	}
	return workspace.ID, nil
}

func (t *TFCloudBackend) initTFEClient() error {
	token, err := t.getToken()
	if err != nil {
		return err
	}
	config := &tfe.Config{
		Token:   token,
		Address: t.opts.TFCloudEndpoint,
	}
	tfcClient, err := tfe.NewClient(config)
	if err != nil {
		return err
	}
	t.client = tfcClient
	return nil
}

func (t *TFCloudBackend) Read(p []byte) (n int, err error) {
	if t.reader == nil {
		if t.client == nil {
			if err := t.initTFEClient(); err != nil {
				return 0, err
			}
		}

		workspaceId, err := t.getWorkspaceId()
		if err != nil {
			return 0, err
		}

		stateVersion, err := t.client.StateVersions.Current(context.Background(), workspaceId)
		if err != nil {
			return 0, errors.Errorf("unable to read current state version: %s", err.Error())
		}

		state, err := t.client.StateVersions.Download(context.Background(), stateVersion.DownloadURL)
		if err != nil {
			return 0, errors.Errorf("unable to download current state content: %s", err.Error())
		}
		t.reader = io.NopCloser(bytes.NewReader(state))
	}
	return t.reader.Read(p)
}

func (t *TFCloudBackend) Close() error {
	if t.reader != nil {
		return t.reader.Close()
	}
	return errors.New("Unable to close reader as nothing was opened")
}
