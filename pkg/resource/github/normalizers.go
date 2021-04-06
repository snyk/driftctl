package github

import "github.com/cloudskiff/driftctl/pkg/resource"

func InitNormalizers() {
	resource.AddNormalizer(GithubBranchProtectionResourceType, githubBranchProtectionNormalizer)
	resource.AddNormalizer(GithubTeamMembershipResourceType, githubTeamMembershipNormalizer)
	resource.AddNormalizer(GithubMembershipResourceType, githubMembershipNormalizer)
	resource.AddNormalizer(GithubTeamResourceType, githubTeamNormalizer)
	resource.AddNormalizer(GithubRepositoryResourceType, githubRepositoryNormalizer)
}
