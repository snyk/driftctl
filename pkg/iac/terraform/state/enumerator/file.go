package enumerator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudskiff/driftctl/pkg/iac/config"
)

type FileEnumeratorConfig struct {
	Bucket *string
	Prefix *string
}

type FileEnumerator struct {
	config config.SupplierConfig
}

func NewFileEnumerator(config config.SupplierConfig) *FileEnumerator {
	return &FileEnumerator{
		config,
	}
}

func (s *FileEnumerator) Path() string {
	return s.config.Path
}

func (s *FileEnumerator) Enumerate() ([]string, error) {
	path := s.config.Path

	info, err := os.Lstat(path)
	if isGlob := HasMeta(path); !isGlob && err != nil {
		return nil, err
	}
	if err == nil {
		// if we got a symlink, use its destination
		if info.Mode()&os.ModeSymlink != 0 {
			destination, err := filepath.EvalSymlinks(path)
			if err != nil {
				return nil, err
			}
			path = destination
			info, err = os.Stat(destination)
			if err != nil {
				return nil, err
			}
		}

		if info != nil && !info.IsDir() {
			return []string{path}, nil
		}

		path = filepath.Join(path, "**/*.tfstate")
	}

	keys, err := Glob(path)
	if err != nil {
		return keys, err
	}

	if len(keys) == 0 {
		return keys, fmt.Errorf("no Terraform state was found in %s, exiting", s.config.Path)
	}

	return keys, err
}
