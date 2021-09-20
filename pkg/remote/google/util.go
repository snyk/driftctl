package google

import (
	"strings"
)

func trimResourceName(name string) string {
	return strings.TrimPrefix(name, "//compute.googleapis.com/")
}
