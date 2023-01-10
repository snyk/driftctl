package enumerator

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

func SplitS3AndGSPath(path string) (prefix string, pattern string) {
	prefix, pattern = splitDirPattern(path)
	return
}

func HasMeta(path string) bool {
	magicChars := `?*[]`
	return strings.ContainsAny(path, magicChars)
}

// Should split a path :
// - prefix : path part that should not contains glob patterns, that is used in S3 query to filter result
// - pattern : should contains the glob pattern to be used by doublestar matching library
func splitDirPattern(p string) (prefix string, pattern string) {
	sep := "/"

	splitPath := strings.Split(p, sep)
	prefixEnded := false
	for _, s := range splitPath {
		if HasMeta(s) || prefixEnded {
			prefixEnded = true
			pattern = strings.Join([]string{pattern, s}, sep)
			continue
		}

		prefix = strings.Join([]string{prefix, s}, sep)

	}
	return strings.Trim(prefix, sep), strings.Trim(pattern, sep)
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
