package enumerator

import (
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/pkg/iac/config"
	"github.com/snyk/driftctl/pkg/iac/terraform/state/backend"
)

type StateEnumerator interface {
	Origin() string
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
