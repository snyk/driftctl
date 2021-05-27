package output

import (
	"fmt"
)

type AttributesGetter interface {
	Attributes() map[string]string
}

func HumanizeAttribute(res AttributesGetter) string {
	if len(res.Attributes()) <= 0 {
		return ""
	}
	attrString := ""
	for key, value := range res.Attributes() {
		if attrString != "" {
			attrString += ", "
		}
		attrString += fmt.Sprintf("%s: %s", key, value)
	}
	return attrString
}
