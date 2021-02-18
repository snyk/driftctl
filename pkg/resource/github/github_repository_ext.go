package github

import "github.com/cloudskiff/driftctl/pkg/resource"

func (r *GithubRepository) NormalizeForState() (resource.Resource, error) {
	if r.Topics != nil && len(*r.Topics) == 0 {
		r.Topics = nil
	}
	return r, nil
}

func (r *GithubRepository) NormalizeForProvider() (resource.Resource, error) {
	if r.Topics != nil && len(*r.Topics) == 0 {
		r.Topics = nil
	}
	return r, nil
}
