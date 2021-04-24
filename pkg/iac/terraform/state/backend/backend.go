package backend

import (
	"fmt"
	"io"
	"net/http"

	"github.com/cloudskiff/driftctl/pkg/iac/config"
	"github.com/pkg/errors"
)

var supportedBackends = []string{
	BackendKeyFile,
	BackendKeyS3,
	BackendKeyHTTP,
	BackendKeyHTTPS,
	BackendKeyCloud,
}

type Backend io.ReadCloser

type Options struct {
	Headers map[string]string
}

func IsSupported(backend string) bool {
	for _, b := range supportedBackends {
		if b == backend {
			return true
		}
	}

	return false
}

func GetBackend(config config.SupplierConfig, opts *Options) (Backend, error) {
	backend := config.Backend

	if !IsSupported(backend) {
		return nil, errors.Errorf("Unsupported backend '%s'", backend)
	}

	switch backend {
	case BackendKeyFile:
		return NewFileReader(config.Path)
	case BackendKeyS3:
		return NewS3Reader(config.Path)
	case BackendKeyHTTP:
		fallthrough
	case BackendKeyHTTPS:
		return NewHTTPReader(&http.Client{}, fmt.Sprintf("%s://%s", config.Backend, config.Path), opts)
	case BackendKeyCloud:
		return NewCloudReader(config.Path, opts)
	default:
		return nil, errors.Errorf("Unsupported backend '%s'", backend)
	}
}

func GetSupportedBackends() []string {
	return supportedBackends[1:]
}
