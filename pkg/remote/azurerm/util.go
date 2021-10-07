package azurerm

import "regexp"

func trimResourceGroupName(name string) string {
	re, _ := regexp.Compile(`\/resourceGroups\/(?P<group>[\w-_]+)\/?`)
	if m := re.FindStringSubmatch(name); len(m) > 0 {
		return m[len(m)-1]
	}
	return ""
}
