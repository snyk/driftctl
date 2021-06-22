package middlewares

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

// Manage tags_all attribute on each compatible resources
type TagsAllManager struct{}

func NewTagsAllManager() TagsAllManager {
	return TagsAllManager{}
}

func (a TagsAllManager) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {
	for _, remoteRes := range *remoteResources {
		if res, ok := remoteRes.(*resource.AbstractResource); ok {
			if _, exist := res.Attrs.Get("tags_all"); exist {
				res.Attrs.SafeDelete([]string{"tags_all"})
			}
		}
	}
	for _, stateRes := range *resourcesFromState {
		if res, ok := stateRes.(*resource.AbstractResource); ok {
			if allTags, exist := res.Attrs.Get("tags_all"); exist {
				_ = res.Attrs.SafeSet([]string{"tags"}, allTags)
				res.Attrs.SafeDelete([]string{"tags_all"})
			}
		}
	}
	return nil
}
