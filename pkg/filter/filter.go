package filter

import "github.com/cloudskiff/driftctl/pkg/resource"

type Filter interface {
	IsResourceIgnored(res resource.Resource) bool
	IsFieldIgnored(res resource.Resource, path []string) bool
}
