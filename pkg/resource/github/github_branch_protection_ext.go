package github

import (
	"encoding/base64"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

func (r *GithubBranchProtection) Attributes() map[string]string {
	attrs := make(map[string]string)
	if r.Pattern != nil {
		repoId := ""
		if r.RepositoryId != nil {
			decodedId, err := base64.StdEncoding.DecodeString(*r.RepositoryId)
			if err == nil {
				repoId = string(decodedId)
			}
		}

		if repoId == "" {
			attrs["Branch"] = *r.Pattern
			attrs["Id"] = r.Id
			return attrs
		}
		attrs["Branch"] = *r.Pattern
		attrs["RepoId"] = repoId
		return attrs
	}
	attrs["Id"] = r.Id
	return attrs
}

func (r *GithubBranchProtection) NormalizeForState() (resource.Resource, error) {
	r.normalize()
	return r, nil
}

func (r *GithubBranchProtection) NormalizeForProvider() (resource.Resource, error) {
	r.normalize()
	return r, nil
}

func (r *GithubBranchProtection) normalize() {
	if r.PushRestrictions != nil && len(*r.PushRestrictions) == 0 {
		r.PushRestrictions = nil
	}
}
