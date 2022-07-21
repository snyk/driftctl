package github

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const GithubTeamMembershipResourceType = "github_team_membership"

func initGithubTeamMembershipMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(GithubTeamMembershipResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"etag"})
	})
	resourceSchemaRepository.SetFlags(GithubTeamMembershipResourceType, resource.FlagDeepMode)
}
