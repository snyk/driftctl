package backend

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"testing"

	"cloud.google.com/go/storage"
	googletest "github.com/snyk/driftctl/test/google"
	"github.com/stretchr/testify/assert"
)

func TestGSBackendInvalid(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *GSBackend
		wantErr error
	}{
		{
			name: "invalid path",
			args: args{
				path: "foobar",
			},
			want:    nil,
			wantErr: fmt.Errorf("Unable to parse Google Storage path: foobar. Must be BUCKET_NAME/PATH/TO/OBJECT"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "testdata/gcp_application_default_credentials.json")

			got, err := NewGSReader(tt.args.path)
			if err.Error() != tt.wantErr.Error() {
				t.Errorf("NewGSReader() error = '%s', wantErr '%s'", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGSReader() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGSBackend_Read(t *testing.T) {
	type args struct {
		bucketName string
		path       string
	}
	tests := []struct {
		name        string
		args        args
		wantErr     error
		handlerFunc map[string]http.HandlerFunc
		expected    string
	}{
		{
			name: "should succeed",
			args: args{
				bucketName: "bucket-1",
				path:       "terraform.tfstate",
			},
			handlerFunc: map[string]http.HandlerFunc{
				"/bucket-1/terraform.tfstate": func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte(`{"version": "1.0.0"}`))
				},
			},
			expected: `{"version": "1.0.0"}`,
		},
		{
			name: "should fail to read remote file",
			args: args{
				bucketName: "bucket-2",
				path:       "path/to/terraform.tfstate",
			},
			wantErr: errors.New("storage: object doesn't exist"),
			handlerFunc: map[string]http.HandlerFunc{
				"/bucket-2/path/to/terraform.tfstate": func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte("Not Found"))
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, server, err := googletest.NewFakeStorageServer(tt.handlerFunc)
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			defer server.Close()

			reader := &GSBackend{
				bucketName:    tt.args.bucketName,
				path:          tt.args.path,
				storageClient: client,
			}
			assert.NoError(t, err)

			got := make([]byte, len(tt.expected))
			_, err = reader.Read(got)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			} else {
				assert.Equal(t, io.EOF, err)
			}
			assert.NotNil(t, got)
			assert.Equal(t, tt.expected, string(got))
		})
	}
}

func TestGSBackend_Close(t *testing.T) {
	tests := []struct {
		name    string
		reader  io.ReadCloser
		client  *storage.Client
		wantErr error
	}{
		{
			name: "should fail to close reader",
			reader: func() io.ReadCloser {
				return nil
			}(),
			client:  &storage.Client{},
			wantErr: errors.New("Unable to close reader as nothing was opened"),
		},
		{
			name: "should close reader",
			reader: func() io.ReadCloser {
				m := &MockReaderMock{}
				m.On("Close").Return(nil)
				return m
			}(),
			client: &storage.Client{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &GSBackend{
				reader:        tt.reader,
				storageClient: tt.client,
			}
			err := h.Close()
			if tt.wantErr == nil {
				assert.Nil(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
			}
		})
	}
}
