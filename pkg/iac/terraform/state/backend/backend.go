package backend

import (
	"fmt"
	"io"

	"github.com/cloudskiff/driftctl/pkg/iac/config"
	"github.com/pkg/errors"
)

var supportedBackends = []string{
	backendFile,
	backendS3,
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
		return nil, errors.New(fmt.Sprintf("Unsupported backend '%s'", backend))
	}

	switch backend {
	case backendFile:
		return NewFileReader(config.Path)
	case backendS3:
		return NewS3Reader(config.Path)
	default:
		return nil, errors.New(fmt.Sprintf("Unsupported backend '%s'", backend))
	}
}

func GetSupportedBackends() []string {
	return supportedBackends[1:]
}
