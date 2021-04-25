package backend

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewCloudReader(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	type args struct {
		workspaceId string
		options     *Options
	}
	tests := []struct {
		name      string
		args      args
		url       string
		wantURL   string
		wantErr   error
		responder httpmock.Responder
	}{
		{
			name: "Should fetch URL with auth header",
			args: args{
				workspaceId: "workspaceId",
				options: &Options{
					Headers: map[string]string{
						"Authorization": "Bearer TOKEN",
					},
				},
			},
			url:       "https://app.terraform.io/api/v2/workspaces/workspaceId/current-state-version",
			wantURL:   "https://archivist.terraform.io/v1/object/test",
			wantErr:   nil,
			responder: httpmock.NewBytesResponder(http.StatusOK, []byte(`{"data":{"attributes":{"hosted-state-download-url":"https://archivist.terraform.io/v1/object/test"}}}`)),
		},
		{
			name: "Should fail with wrong workspaceId",
			args: args{
				workspaceId: "wrong_workspaceId",
				options: &Options{
					Headers: map[string]string{
						"Authorization": "Bearer TOKEN",
					},
				},
			},
			url:       "https://app.terraform.io/api/v2/workspaces/wrong_workspaceId/current-state-version",
			wantURL:   "",
			wantErr:   errors.New("Error reading state from Terraform Cloud/Enterprise workspace: wrong workspace id"),
			responder: httpmock.NewBytesResponder(http.StatusNotFound, []byte{}),
		},
		{
			name: "Should fail with bad authentication token",
			args: args{
				workspaceId: "workspaceId",
				options: &Options{
					Headers: map[string]string{
						"Authorization": "Bearer WRONG_TOKEN",
					},
				},
			},
			url:       "https://app.terraform.io/api/v2/workspaces/workspaceId/current-state-version",
			wantURL:   "",
			wantErr:   errors.New("Error reading state from Terraform Cloud/Enterprise workspace: bad authentication token"),
			responder: httpmock.NewBytesResponder(http.StatusUnauthorized, []byte{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Reset()
			httpmock.RegisterResponder("GET", tt.url, tt.responder)
			if tt.name == "Should fetch URL with auth header" {
				httpmock.RegisterResponder("GET", "https://archivist.terraform.io/v1/object/test", httpmock.NewBytesResponder(http.StatusOK, []byte(`{}`)))
			}
			got, err := NewCloudReader(tt.args.workspaceId, tt.args.options)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			} else {
				assert.NoError(t, err)
			}
			assert.NotNil(t, got)
			assert.Equal(t, tt.wantURL, got.url)
		})
	}
}
