package enumerator

import (
	"path"
	"strings"
)

func SplitPath(path string) (prefix string, pattern string) {
	prefix, pattern = extractPathSegments(path)
	return
}

// Extracts and returns the below segments:
// - prefix : path part that should not contains glob patterns, that is used in S3 query to filter result
// - pattern : should contains the glob pattern to be used by doublestar matching library
func extractPathSegments(p string) (prefix string, pattern string) {
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

// creates a path that includes both the prefix and the glob pattern (if any is present)
func JoinAndTrimPath(prefix string, pattern string) string {
	return path.Join(prefix, pattern)
}

// HasMeta reports whether path contains any of the magic characters
func HasMeta(path string) bool {
	magicChars := `?*[]`
	return strings.ContainsAny(path, magicChars)
}
