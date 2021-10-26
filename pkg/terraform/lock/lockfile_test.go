package lock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ReadLockFile(t *testing.T) {
	cases := []struct {
		test     string
		filepath string
		assert   func(*testing.T, *Lockfile, error)
	}{
		{
			test:     "should attempt to read non existing file",
			filepath: "testdata/file_does_not_exist.hcl",
			assert: func(tt *testing.T, locks *Lockfile, err error) {
				provider := locks.GetProviderByAddress(&ProviderAddress{
					Type:      "aws",
					Namespace: "hashicorp",
					Hostname:  "registry.terraform.io",
				})

				assert.Len(tt, locks.Providers, 0)
				assert.Nil(tt, provider)
				assert.EqualError(tt, err, "<nil>: Failed to read file; The configuration file \"testdata/file_does_not_exist.hcl\" could not be read.")
			},
		},
		{
			test:     "should read valid lock file",
			filepath: "testdata/lockfile_valid.hcl",
			assert: func(tt *testing.T, locks *Lockfile, err error) {
				provider := locks.GetProviderByAddress(&ProviderAddress{
					Type:      "aws",
					Namespace: "hashicorp",
					Hostname:  "registry.terraform.io",
				})

				assert.Len(tt, locks.Providers, 10)
				assert.Equal(tt, "3.47.0", provider.Version)
				assert.Equal(tt, "registry.terraform.io/hashicorp/aws", provider.Address)
				assert.Equal(tt, "~> 3.47.0", provider.Constraints)
				assert.Equal(tt, false, err != nil)
			},
		},
		{
			test:     "should fail to retrieve provider block with invalid address",
			filepath: "testdata/lockfile_valid.hcl",
			assert: func(tt *testing.T, locks *Lockfile, err error) {
				provider := locks.GetProviderByAddress(&ProviderAddress{})

				assert.Len(tt, locks.Providers, 10)
				assert.Nil(tt, provider)
				assert.Nil(tt, err)
			},
		},
		{
			test:     "should read empty file without error",
			filepath: "testdata/lockfile_empty.hcl",
			assert: func(tt *testing.T, locks *Lockfile, err error) {
				provider := locks.GetProviderByAddress(&ProviderAddress{})

				assert.Len(tt, locks.Providers, 0)
				assert.Nil(tt, provider)
				assert.Nil(tt, err)
			},
		},
		{
			test:     "should return error for invalid lock file",
			filepath: "testdata/lockfile_invalid.hcl",
			assert: func(tt *testing.T, locks *Lockfile, err error) {
				provider := locks.GetProviderByAddress(&ProviderAddress{})

				assert.Len(tt, locks.Providers, 1)
				assert.Nil(tt, provider)
				assert.EqualError(tt, err, "testdata/lockfile_invalid.hcl:4,48-48: Missing required argument; The argument \"version\" is required, but no definition was found.")
			},
		},
		{
			test:     "should parse provider blocks without error",
			filepath: "testdata/lockfile_invalid_type-1.hcl",
			assert: func(tt *testing.T, locks *Lockfile, err error) {
				provider := locks.GetProviderByAddress(&ProviderAddress{
					Type:      "google-beta",
					Namespace: "hashicorp",
					Hostname:  "registry.terraform.io",
				})

				assert.Len(tt, locks.Providers, 2)
				assert.Equal(tt, "2.71.0", provider.Version)
				assert.Equal(tt, "registry.terraform.io/hashicorp/google-beta", provider.Address)
				assert.Equal(tt, "~> 2.71.0", provider.Constraints)
				assert.Nil(tt, err)
			},
		},
		{
			test:     "should parse provider blocks without error",
			filepath: "testdata/lockfile_invalid_type-3.hcl",
			assert: func(tt *testing.T, locks *Lockfile, err error) {
				provider := locks.GetProviderByAddress(&ProviderAddress{
					Type:      "google-beta",
					Namespace: "hashicorp",
					Hostname:  "registry.terraform.io",
				})

				assert.Len(tt, locks.Providers, 2)
				assert.Equal(tt, "2.71.0", provider.Version)
				assert.Equal(tt, "registry.terraform.io/hashicorp/google-beta", provider.Address)
				assert.Equal(tt, "~> 2.71.0", provider.Constraints)
				assert.Nil(tt, err)
			},
		},
		{
			test:     "should not find provider address",
			filepath: "testdata/lockfile_valid.hcl",
			assert: func(tt *testing.T, locks *Lockfile, err error) {
				provider := locks.GetProviderByAddress(&ProviderAddress{
					Type:      "unknown",
					Namespace: "hashicorp",
					Hostname:  "registry.terraform.io",
				})

				assert.Len(tt, locks.Providers, 10)
				assert.Nil(tt, provider)
				assert.Nil(tt, err)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.test, func(tt *testing.T) {
			locks, err := ReadLocksFromFile(c.filepath)
			c.assert(t, locks, err)
		})
	}
}
