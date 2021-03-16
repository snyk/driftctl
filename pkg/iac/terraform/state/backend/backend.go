package backend

import (
	"fmt"
	"io"

	"github.com/cloudskiff/driftctl/pkg/iac/config"
	"github.com/pkg/errors"
)

var supportedBackends = []string{
	BackendKeyFile,
	BackendKeyS3,
	BackendKeyHTTP,
	BackendKeyHTTPS,
}

type Backend io.ReadCloser

func IsSupported(backend string) bool {
	for _, b := range supportedBackends {
		if b == backend {
			return true
		}
	}

	return false
}

func GetBackend(config config.SupplierConfig) (Backend, error) {

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
		return NewHTTPReader(fmt.Sprintf("%s://%s", config.Backend, config.Path))
	default:
		return nil, errors.Errorf("Unsupported backend '%s'", backend)
	}
}

func GetSupportedBackends() []string {
	return supportedBackends[1:]
}
