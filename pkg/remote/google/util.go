package google

import (
	"regexp"
)

func trimResourceName(name string) string {
	re, _ := regexp.Compile(`^\/\/[\w]+.googleapis.com\/`)
	return re.ReplaceAllString(name, "")
}
