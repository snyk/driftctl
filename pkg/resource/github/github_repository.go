package github

import "github.com/cloudskiff/driftctl/pkg/resource"

const GithubRepositoryResourceType = "github_repository"

func initGithubRepositoryMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(GithubRepositoryResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"auto_init"})
		val.SafeDelete([]string{"etag"})
	})
	resourceSchemaRepository.SetFlags(GithubRepositoryResourceType, resource.FlagDeepMode)
}
