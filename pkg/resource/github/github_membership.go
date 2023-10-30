package github

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const GithubMembershipResourceType = "github_membership"

func initGithubMembershipMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(GithubMembershipResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"etag"})
	})
}
