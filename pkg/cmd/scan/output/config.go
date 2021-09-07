package output

import "fmt"

type OutputConfig struct {
	Key  string
	Path string
}

func (o *OutputConfig) String() string {
	return fmt.Sprintf("%s://%s", o.Key, o.Path)
}
