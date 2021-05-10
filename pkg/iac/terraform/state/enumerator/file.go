package enumerator

import (
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

	return keys, err
}
