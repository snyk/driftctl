package github

import (
	"github.com/cloudskiff/driftctl/pkg/dctlcty"
)

func InitMetadatas() {
	dctlcty.SetMetadata(GithubBranchProtectionResourceType, githubBranchProtectionTags, githubBranchProtectionNormalizer)
	dctlcty.SetMetadata(GithubTeamMembershipResourceType, githubTeamMembershipTags, githubTeamMembershipNormalizer)
	dctlcty.SetMetadata(GithubMembershipResourceType, githubMembershipTags, githubMembershipNormalizer)
	dctlcty.SetMetadata(GithubTeamResourceType, githubTeamTags, githubTeamNormalizer)
	dctlcty.SetMetadata(GithubRepositoryResourceType, githubRepositoryTags, githubRepositoryNormalizer)
}
