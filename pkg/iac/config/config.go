package config

import "fmt"

type SupplierConfig struct {
	Key     string
	Backend string
	Path    string
}

func (c *SupplierConfig) String() string {
	str := c.Key
	if c.Backend != "" {
		str += fmt.Sprintf("+%s", c.Backend)
	}
	if str != "" {
		str += "://"
	}
	if c.Path != "" {
		str += c.Path
	}
	return str
}
