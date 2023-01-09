package enumerator

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/snyk/driftctl/pkg/iac/config"
	googletest "github.com/snyk/driftctl/test/google"
	"github.com/stretchr/testify/assert"
)

func TestGSEnumerator_NewGSEnumerator(t *testing.T) {
	tests := []struct {
		name   string
		config config.SupplierConfig
	}{
		{
			name: "test if required attributes are not nil",
			config: config.SupplierConfig{
				Key:     "tfstate",
				Backend: "gs",
				Path:    "terraform.tfstate",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGSEnumerator(tt.config)
			assert.NotNil(t, got)
			assert.NotNil(t, got.client)
			assert.NotNil(t, got.config)
		})
	}
}

func TestGSEnumerator_NewGSEnumerator_HasCorrectConfig(t *testing.T) {
	tests := []struct {
		name   string
		config config.SupplierConfig
	}{
		{
			name: "test whether the supplier config isn't changed",
			config: config.SupplierConfig{
				Key:     "tfstate",
				Backend: "gs",
				Path:    "terraform.tfstate",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGSEnumerator(tt.config)
			assert.NotNil(t, got)
			assert.Equal(t, tt.config.Key, got.config.Key)
			assert.Equal(t, tt.config.Backend, got.config.Backend)
			assert.Equal(t, tt.config.Path, got.config.Path)
		})
	}
}

func TestGSEnumerator_Enumerate(t *testing.T) {
	tests := []struct {
		name        string
		config      config.SupplierConfig
		handlerFunc map[string]http.HandlerFunc
		want        []string
		err         string
	}{
		{
			name: "should succeed",
			config: config.SupplierConfig{
				Path: "bucket-1/*.tfstate",
			},
			handlerFunc: map[string]http.HandlerFunc{
				"/bucket-2/path/to/terraform.tfstate": func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("Not Found"))
				},
			},
			want: []string{
				"bucket-1/a/nested/prefix/1/state1.tfstate",
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

			gs := &GSEnumerator{
				config: tt.config,
				client: *client,
			}

			assert.NoError(t, err)

			got, err := gs.Enumerate()
			if err != nil && err.Error() != tt.err {
				t.Fatalf("Expected error '%s', got '%s'", tt.err, err.Error())
				return
			}
			if tt.err != "" && err == nil {
				t.Fatalf("Expected error '%s' but got nil", tt.err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Enumerate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
