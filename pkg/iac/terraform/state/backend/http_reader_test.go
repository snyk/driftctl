package backend

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	pkghttp "github.com/cloudskiff/driftctl/pkg/http"
	"github.com/stretchr/testify/assert"
)

func TestHTTPBackend_Read(t *testing.T) {
	type args struct {
		url     string
		options *Options
	}
	tests := []struct {
		name       string
		args       args
		wantErr    error
		httpClient pkghttp.HTTPClient
		expected   string
	}{
		{
			name: "Should fail with wrong URL",
			args: args{
				url: "wrong_url",
				options: &Options{
					Headers: map[string]string{},
				},
			},
			wantErr: errors.New("Get \"wrong_url\": unsupported protocol scheme \"\""),
			httpClient: func() pkghttp.HTTPClient {
				return &http.Client{}
			}(),
			expected: "",
		},
		{
			name: "Should fetch URL with auth header",
			args: args{
				url: "https://example.com/cloudskiff/driftctl/main/terraform.tfstate",
				options: &Options{
					Headers: map[string]string{
						"Authorization": "Basic Test",
					},
				},
			},
			wantErr: nil,
			httpClient: func() pkghttp.HTTPClient {
				m := &pkghttp.MockHTTPClient{}

				req, _ := http.NewRequest(http.MethodGet, "https://example.com/cloudskiff/driftctl/main/terraform.tfstate", nil)

				req.Header.Add("Authorization", "Basic Test")

				bodyReader := strings.NewReader("{}")
				bodyReadCloser := io.NopCloser(bodyReader)

				m.On("Do", req).Return(&http.Response{
					StatusCode: 200,
					Body:       bodyReadCloser,
				}, nil)

				return m
			}(),
			expected: "{}",
		},
		{
			name: "Should fail with bad status code",
			args: args{
				url: "https://example.com/cloudskiff/driftctl/main/terraform.tfstate",
				options: &Options{
					Headers: map[string]string{},
				},
			},
			wantErr: errors.New("error requesting HTTP(s) backend state: status code: 404"),
			httpClient: func() pkghttp.HTTPClient {
				m := &pkghttp.MockHTTPClient{}

				req, _ := http.NewRequest(http.MethodGet, "https://example.com/cloudskiff/driftctl/main/terraform.tfstate", nil)

				bodyReader := strings.NewReader("test")
				bodyReadCloser := io.NopCloser(bodyReader)

				m.On("Do", req).Return(&http.Response{
					StatusCode: 404,
					Body:       bodyReadCloser,
				}, nil)

				return m
			}(),
			expected: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := NewHTTPReader(tt.httpClient, tt.args.url, tt.args.options)
			assert.NoError(t, err)

			got := make([]byte, len(tt.expected))
			_, err = reader.Read(got)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			} else {
				assert.NoError(t, err)
			}
			assert.NotNil(t, got)
			assert.Equal(t, tt.expected, string(got))
		})
	}
}

func TestHTTPBackend_Close(t *testing.T) {
	type fields struct {
		req    *http.Request
		reader io.ReadCloser
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "should fail to close reader",
			fields: fields{
				req: &http.Request{},
				reader: func() io.ReadCloser {
					return nil
				}(),
			},
			wantErr: true,
		},
		{
			name: "should close reader",
			fields: fields{
				req: &http.Request{},
				reader: func() io.ReadCloser {
					m := &MockReaderMock{}
					m.On("Close").Return(nil)
					return m
				}(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HTTPBackend{
				request: tt.fields.req,
				reader:  tt.fields.reader,
			}
			if err := h.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
