package github

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const GithubTeamMembershipResourceType = "github_team_membership"

func initGithubTeamMembershipMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(GithubTeamMembershipResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"etag"})
	})
}
