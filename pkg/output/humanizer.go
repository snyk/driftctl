package output

import (
	"fmt"
	"sort"
)

type AttributesGetter interface {
	HumanReadableAttributes() map[string]string
}

func HumanizeAttribute(res AttributesGetter) string {
	attributes := res.HumanReadableAttributes()
	if len(attributes) <= 0 {
		return ""
	}
	// sort attributes
	keys := make([]string, 0, len(attributes))
	for k := range attributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// retrieve stringer
	attrString := ""
	for _, k := range keys {
		if attrString != "" {
			attrString += ", "
		}
		attrString += fmt.Sprintf("%s: %s", k, attributes[k])
	}
	return attrString
}
