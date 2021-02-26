package github

import (
	"encoding/base64"
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

func (r *GithubBranchProtection) String() string {
	if r.Pattern != nil {
		repoId := ""
		if r.RepositoryId != nil {
			decodedId, err := base64.StdEncoding.DecodeString(*r.RepositoryId)
			if err == nil {
				repoId = string(decodedId)
			}
		}

		if repoId == "" {
			return fmt.Sprintf("Branch: %s (Id: %s)", *r.Pattern, r.Id)
		}
		return fmt.Sprintf("Branch: %s (RepoId: %s)", *r.Pattern, repoId)
	}
	return r.Id
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
