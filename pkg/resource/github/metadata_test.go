package github_test

import (
	"testing"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/github"
	testresource "github.com/snyk/driftctl/test/resource"
	"github.com/stretchr/testify/assert"
)

func TestGitHub_Metadata_Flags(t *testing.T) {
	testcases := map[string][]resource.Flags{
		github.GithubBranchProtectionResourceType: {},
		github.GithubMembershipResourceType:       {},
		github.GithubTeamMembershipResourceType:   {},
		github.GithubRepositoryResourceType:       {},
		github.GithubTeamResourceType:             {},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("github", "4.4.0")
	github.InitResourcesMetadata(schemaRepository)

	for ty, flags := range testcases {
		t.Run(ty, func(tt *testing.T) {
			sch, exist := schemaRepository.GetSchema(ty)
			assert.True(tt, exist)

			if len(flags) == 0 {
				assert.Equal(tt, resource.Flags(0x0), sch.Flags, "should not have any flag")
				return
			}

			for _, flag := range flags {
				assert.Truef(tt, sch.Flags.HasFlag(flag), "should have given flag %d", flag)
			}
		})
	}
}
