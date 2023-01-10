package enumerator

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

func Glob(pattern string) ([]string, error) {
	if !strings.Contains(pattern, "**") {
		return filepath.Glob(pattern)
	}

	var files []string

	err := doublestar.GlobWalk(os.DirFS("."), path.Clean(pattern), func(path string, d fs.DirEntry) error {
		// Ensure paths aren't actually directories
		// For example when the directory matches the glob pattern like it's a file
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}
