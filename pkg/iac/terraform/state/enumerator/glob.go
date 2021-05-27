package enumerator

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

func GlobS3(path string) (prefix string, pattern string) {
	if !HasMeta(path) {
		return path, ""
	}
	prefix, pattern = splitDirPattern(path)
	return
}

func HasMeta(path string) bool {
	magicChars := `?*[]`
	return strings.ContainsAny(path, magicChars)
}

func splitDirPattern(p string) (base string, pattern string) {
	base = p
	sep := string(os.PathSeparator)

	for {
		if !HasMeta(base) {
			break
		}
		if !strings.Contains(base, sep) {
			return "", base
		}
		base = base[:strings.LastIndex(base, sep)]
	}
	if len(base) == len(p) {
		return p, ""
	}
	return base, p[len(base)+1:]
}

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
