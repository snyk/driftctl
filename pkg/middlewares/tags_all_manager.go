package middlewares

import (
	"github.com/snyk/driftctl/pkg/resource"
)

// Manage tags_all attribute on each compatible resources
type TagsAllManager struct{}

func NewTagsAllManager() TagsAllManager {
	return TagsAllManager{}
}

func (a TagsAllManager) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	for _, remoteRes := range *remoteResources {
		if remoteRes.Attrs != nil {
			if _, exist := remoteRes.Attrs.Get("tags_all"); exist {
				remoteRes.Attrs.SafeDelete([]string{"tags_all"})
			}
		}
	}
	for _, stateRes := range *resourcesFromState {
		if stateRes.Attrs != nil {
			if allTags, exist := stateRes.Attrs.Get("tags_all"); exist {
				_ = stateRes.Attrs.SafeSet([]string{"tags"}, allTags)
				stateRes.Attrs.SafeDelete([]string{"tags_all"})
			}
		}
	}
	return nil
}
