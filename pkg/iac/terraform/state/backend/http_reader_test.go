package backend

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/mocks"
)

func TestNewHTTPReader(t *testing.T) {
	type args struct {
		url     string
		options *Options
	}
	tests := []struct {
		name       string
		args       args
		wantErr    error
		httpClient func() HttpClient
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
			httpClient: func() HttpClient {
				return &http.Client{}
			},
			expected: "",
		},
		{
			name: "Should fetch URL with auth header",
			args: args{
				url: "https://wrong.url/cloudskiff/driftctl/main/terraform.tfstate",
				options: &Options{
					Headers: map[string]string{
						"Authorization": "Basic Test",
					},
				},
			},
			wantErr: nil,
			httpClient: func() HttpClient {
				m := &mocks.HttpClient{}

				req, _ := http.NewRequest(http.MethodGet, "https://wrong.url/cloudskiff/driftctl/main/terraform.tfstate", nil)

				req.Header.Add("Authorization", "Basic Test")

				bodyReader := strings.NewReader("{}")
				bodyReadCloser := io.NopCloser(bodyReader)

				m.On("Do", req).Return(&http.Response{
					StatusCode: 200,
					Body:       bodyReadCloser,
				}, nil)

				return m
			},
			expected: "{}",
		},
		{
			name: "Should fail with bad status code",
			args: args{
				url: "https://wrong.url/cloudskiff/driftctl/main/terraform.tfstate",
				options: &Options{
					Headers: map[string]string{},
				},
			},
			wantErr: errors.New("error requesting HTTP(s) backend state: status code: 404"),
			httpClient: func() HttpClient {
				m := &mocks.HttpClient{}

				req, _ := http.NewRequest(http.MethodGet, "https://wrong.url/cloudskiff/driftctl/main/terraform.tfstate", nil)

				bodyReader := strings.NewReader("test")
				bodyReadCloser := io.NopCloser(bodyReader)

				m.On("Do", req).Return(&http.Response{
					StatusCode: 404,
					Body:       bodyReadCloser,
				}, nil)

				return m
			},
			expected: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHTTPReader(tt.httpClient(), tt.args.url, tt.args.options)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			} else {
				assert.NoError(t, err)
			}
			assert.NotNil(t, got)
			gotBytes, err := io.ReadAll(got)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, string(gotBytes))
		})
	}
}

func TestHTTPBackend_Close(t *testing.T) {
	type fields struct {
		url    string
		reader func() io.ReadCloser
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "should fail to close reader",
			fields: fields{
				url: "",
				reader: func() io.ReadCloser {
					return nil
				},
			},
			wantErr: true,
		},
		{
			name: "should close reader",
			fields: fields{
				url: "",
				reader: func() io.ReadCloser {
					m := &MockReaderMock{}
					m.On("Close").Return(nil)
					return m
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HTTPBackend{
				url:    tt.fields.url,
				reader: tt.fields.reader(),
			}
			if err := h.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHTTPBackend_Read(t *testing.T) {
	type fields struct {
		url    string
		reader func() io.ReadCloser
	}
	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantN   int
		wantErr error
	}{
		{
			name: "should fail to read because of nil reader",
			fields: fields{
				url: "",
				reader: func() io.ReadCloser {
					return nil
				},
			},
			wantErr: errors.New("Reader not initialized"),
		},
		{
			name: "should fail to read",
			fields: fields{
				url: "",
				reader: func() io.ReadCloser {
					m := &MockReaderMock{}
					m.On("Read", mock.Anything).Return(0, errors.New("test"))
					return m
				},
			},
			wantErr: errors.New("test"),
		},
		{
			name: "should read",
			fields: fields{
				url: "",
				reader: func() io.ReadCloser {
					m := &MockReaderMock{}
					m.On("Read", mock.Anything).Return(0, nil)
					return m
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HTTPBackend{
				url:    tt.fields.url,
				reader: tt.fields.reader(),
			}
			gotN, err := h.Read(tt.args.p)

			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			if gotN != tt.wantN {
				t.Errorf("Read() gotN = %v, want %v", gotN, tt.wantN)
			}
		})
	}
}
