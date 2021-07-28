package enumerator

import (
	"github.com/cloudskiff/driftctl/pkg/iac/config"
	"github.com/cloudskiff/driftctl/pkg/iac/terraform/state/backend"
	"github.com/sirupsen/logrus"
)

type StateEnumerator interface {
	Path() string
	Enumerate() ([]string, error)
}

func GetEnumerator(config config.SupplierConfig) StateEnumerator {

	switch config.Backend {
	case backend.BackendKeyFile:
		return NewFileEnumerator(config)
	case backend.BackendKeyS3:
		return NewS3Enumerator(config)
	}

	logrus.WithFields(logrus.Fields{
		"backend": config.Backend,
	}).Debug("No enumerator for backend")

	return nil
}
