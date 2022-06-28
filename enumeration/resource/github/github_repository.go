package github

import "github.com/snyk/driftctl/enumeration/resource"

const GithubRepositoryResourceType = "github_repository"

func initGithubRepositoryMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(GithubRepositoryResourceType, resource.FlagDeepMode)
}
