package filter

import "github.com/cloudskiff/driftctl/pkg/resource"

type Filter interface {
	IsTypeIgnored(ty resource.ResourceType) bool
	IsResourceIgnored(res resource.Resource) bool
	IsFieldIgnored(res resource.Resource, path []string) bool
}
