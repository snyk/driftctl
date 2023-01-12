package enumerator

import (
	"strings"
)

// Returns the below segments:
// - prefix : path part that should not contains glob patterns, that is used in S3 query to filter result
// - pattern : should contains the glob pattern to be used by doublestar matching library
func extractPrefixAndPattern(path string) (prefix string, pattern string) {
	sep := "/"

	splitPath := strings.Split(path, sep)
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

// HasMeta reports whether path contains any of the magic characters
func HasMeta(path string) bool {
	magicChars := `?*[]`
	return strings.ContainsAny(path, magicChars)
}
