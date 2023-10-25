package github

import (
	"encoding/base64"

	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const GithubBranchProtectionResourceType = "github_branch_protection"

func initGithubBranchProtectionMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(GithubBranchProtectionResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"repository_id"}) // Terraform provider is always returning nil
	})
	resourceSchemaRepository.SetHumanReadableAttributesFunc(GithubBranchProtectionResourceType, func(res *resource.Resource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		pattern := val.GetString("pattern")
		repoID := val.GetString("repository_id")
		if pattern != nil && *pattern != "" {
			id := ""
			if repoID != nil && *repoID != "" {
				decodedID, err := base64.StdEncoding.DecodeString(*repoID)
				if err == nil {
					id = string(decodedID)
				}
			}
			if id == "" {
				attrs["Branch"] = *pattern
				attrs["Id"] = res.ResourceId()
				return attrs
			}
			attrs["Branch"] = *pattern
			attrs["RepoId"] = id
			return attrs
		}
		attrs["Id"] = res.ResourceId()
		return attrs
	})
}
