package state

import (
	"fmt"

	"github.com/hashicorp/go-version"
)

var (
	// UnsupportedVersionConstraints is an array of version constraints known to be unsupported.
	// If a given state matches one of these, all resources of the related state will be ignored and marked as drifted.
	UnsupportedVersionConstraints = []string{"<0.11.0"}
)

type UnsupportedVersionError struct {
	StateFile string
	Version   *version.Version
}

func (u *UnsupportedVersionError) Error() string {
	return fmt.Sprintf("%s was generated using Terraform %s which is currently not supported by driftctl. Please read documentation at https://docs.driftctl.com/limitations", u.StateFile, u.Version)
}

func IsVersionSupported(rawVersion string) (bool, error) {
	v, err := version.NewVersion(rawVersion)
	if err != nil {
		return false, err
	}

	for _, rawConstraint := range UnsupportedVersionConstraints {
		c, err := version.NewConstraint(rawConstraint)
		if err != nil {
			return false, err
		}

		if c.Check(v) {
			return false, nil
		}
	}

	return true, nil
}
