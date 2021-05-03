package backend

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewTFCloudReader(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	type args struct {
		workspaceId string
		options     *Options
	}
	tests := []struct {
		name    string
		args    args
		url     string
		wantURL string
		wantErr error
		mock    func()
	}{
		{
			name: "Should fetch URL with auth header",
			args: args{
				workspaceId: "workspaceId",
				options: &Options{
					TFCloudToken: "TOKEN",
				},
			},
			url:     "https://app.terraform.io/api/v2/workspaces/workspaceId/current-state-version",
			wantURL: "https://archivist.terraform.io/v1/object/test",
			wantErr: nil,
			mock: func() {
				httpmock.Reset()
				httpmock.RegisterResponder(
					"GET",
					"https://app.terraform.io/api/v2/workspaces/workspaceId/current-state-version",
					httpmock.NewBytesResponder(http.StatusOK, []byte(`{"data":{"attributes":{"hosted-state-download-url":"https://archivist.terraform.io/v1/object/test"}}}`)),
				)
				httpmock.RegisterResponder(
					"GET",
					"https://archivist.terraform.io/v1/object/test",
					httpmock.NewBytesResponder(http.StatusOK, []byte(`{}`)),
				)
			},
		},
		{
			name: "Should fail with wrong workspaceId",
			args: args{
				workspaceId: "wrong_workspaceId",
				options: &Options{
					TFCloudToken: "TOKEN",
				},
			},
			url:     "https://app.terraform.io/api/v2/workspaces/wrong_workspaceId/current-state-version",
			wantURL: "",
			mock: func() {
				httpmock.Reset()
				httpmock.RegisterResponder(
					"GET",
					"https://app.terraform.io/api/v2/workspaces/wrong_workspaceId/current-state-version",
					httpmock.NewBytesResponder(http.StatusNotFound, []byte{}),
				)
			},
			wantErr: errors.New("error requesting terraform cloud backend state: status code: 404"),
		},
		{
			name: "Should fail with bad authentication token",
			args: args{
				workspaceId: "workspaceId",
				options: &Options{
					TFCloudToken: "TOKEN",
				},
			},
			url:     "https://app.terraform.io/api/v2/workspaces/workspaceId/current-state-version",
			wantURL: "",
			mock: func() {
				httpmock.Reset()
				httpmock.RegisterResponder(
					"GET",
					"https://app.terraform.io/api/v2/workspaces/workspaceId/current-state-version",
					httpmock.NewBytesResponder(http.StatusUnauthorized, []byte{}),
				)
			},
			wantErr: errors.New("error requesting terraform cloud backend state: status code: 401"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := NewTFCloudReader(&http.Client{}, tt.args.workspaceId, tt.args.options)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			} else {
				assert.NoError(t, err)
			}
			assert.NotNil(t, got)
			assert.Equal(t, tt.wantURL, got.request.URL.String())
		})
	}
}
