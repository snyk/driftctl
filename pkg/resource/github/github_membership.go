package github

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const GithubMembershipResourceType = "github_membership"

func initGithubMembershipMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(GithubMembershipResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"etag"})
	})
	resourceSchemaRepository.SetFlags(GithubMembershipResourceType, resource.FlagDeepMode)
}
