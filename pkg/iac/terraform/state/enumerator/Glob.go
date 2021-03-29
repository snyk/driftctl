package enumerator

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func GlobS3(path string) (prefix string, pattern string, err error) {
	if !HasMeta(path) {
		return path, "", nil
	}
	if strings.Contains(path, "**") {
		return prefix, pattern, errors.New("** not supported for S3 pattern")
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

	globs := strings.Split(pattern, "**")
	var files = []string{""}

	for _, glob := range globs {
		var matches []string
		var exists = map[string]bool{}
		for _, match := range files {
			paths, err := filepath.Glob(match + glob)
			if err != nil {
				return nil, err
			}
			for _, path := range paths {
				err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if _, ok := exists[path]; !ok {
						matches = append(matches, path)
						exists[path] = true
					}
					return nil
				})
				if err != nil {
					return nil, err
				}
			}
		}
		files = matches
	}

	return files, nil
}
