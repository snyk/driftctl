package state

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestUnsupportedVersions(t *testing.T) {
	var tests = []struct {
		name        string
		constraints []string
		version     string
		supported   bool
		err         error
	}{
		{
			name:        "should not support 0.13",
			constraints: []string{"=0.13"},
			version:     "0.13.0",
			supported:   false,
			err:         nil,
		},
		{
			name:        "should return error on version parsing",
			constraints: UnsupportedVersionConstraints,
			version:     "test",
			supported:   false,
			err:         errors.New("Malformed version: test"),
		},
		{
			name:        "should return error on constraint parsing",
			constraints: []string{"bad_constraint"},
			version:     "0.14",
			supported:   false,
			err:         errors.New("Malformed constraint: bad_constraint"),
		},
		{
			name:        "should support 0.11",
			constraints: UnsupportedVersionConstraints,
			version:     "0.11.0",
			supported:   true,
			err:         nil,
		},
		{
			name:        "should support 0.14.8",
			constraints: UnsupportedVersionConstraints,
			version:     "0.14.8",
			supported:   true,
			err:         nil,
		},
		{
			name:        "should not support <0.11",
			constraints: UnsupportedVersionConstraints,
			version:     "0.10.9",
			supported:   false,
			err:         nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			UnsupportedVersionConstraints = test.constraints

			got, err := IsVersionSupported(test.version)
			assert.Equal(t, test.supported, got)

			if err != nil {
				assert.EqualError(t, test.err, err.Error())
			} else {
				assert.Equal(t, test.err, err)
			}
		})
	}
}
