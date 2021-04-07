package github

import (
	"github.com/zclconf/go-cty/cty"

	"github.com/cloudskiff/driftctl/pkg/dctlcty"
)

const GithubRepositoryResourceType = "github_repository"

type GithubRepository struct {
	AllowMergeCommit    *bool     `cty:"allow_merge_commit"`
	AllowRebaseMerge    *bool     `cty:"allow_rebase_merge"`
	AllowSquashMerge    *bool     `cty:"allow_squash_merge"`
	ArchiveOnDestroy    *bool     `cty:"archive_on_destroy"`
	Archived            *bool     `cty:"archived"`
	AutoInit            *bool     `cty:"auto_init" diff:"-"`
	DefaultBranch       *string   `cty:"default_branch" computed:"true"`
	DeleteBranchOnMerge *bool     `cty:"delete_branch_on_merge"`
	Description         *string   `cty:"description"`
	Etag                *string   `cty:"etag" diff:"-"`
	FullName            *string   `cty:"full_name" computed:"true"`
	GitCloneUrl         *string   `cty:"git_clone_url" computed:"true"`
	GitignoreTemplate   *string   `cty:"gitignore_template"`
	HasDownloads        *bool     `cty:"has_downloads"`
	HasIssues           *bool     `cty:"has_issues"`
	HasProjects         *bool     `cty:"has_projects"`
	HasWiki             *bool     `cty:"has_wiki"`
	HomepageUrl         *string   `cty:"homepage_url"`
	HtmlUrl             *string   `cty:"html_url" computed:"true"`
	HttpCloneUrl        *string   `cty:"http_clone_url" computed:"true"`
	Id                  string    `cty:"id" computed:"true"`
	IsTemplate          *bool     `cty:"is_template"`
	LicenseTemplate     *string   `cty:"license_template"`
	Name                *string   `cty:"name"`
	NodeId              *string   `cty:"node_id" computed:"true"`
	Private             *bool     `cty:"private" computed:"true"`
	RepoId              *int      `cty:"repo_id" computed:"true"`
	SshCloneUrl         *string   `cty:"ssh_clone_url" computed:"true"`
	SvnUrl              *string   `cty:"svn_url" computed:"true"`
	Topics              *[]string `cty:"topics"` // Can be null after a tf apply
	Visibility          *string   `cty:"visibility" computed:"true"`
	VulnerabilityAlerts *bool     `cty:"vulnerability_alerts"`
	Pages               *[]struct {
		Cname     *string `cty:"cname"`
		Custom404 *bool   `cty:"custom_404" computed:"true"`
		HtmlUrl   *string `cty:"html_url" computed:"true"`
		Status    *string `cty:"status" computed:"true"`
		Url       *string `cty:"url" computed:"true"`
		Source    *[]struct {
			Branch *string `cty:"branch"`
			Path   *string `cty:"path"`
		} `cty:"source"`
	} `cty:"pages"`
	Template *[]struct {
		Owner      *string `cty:"owner"`
		Repository *string `cty:"repository"`
	} `cty:"template"`
	CtyVal *cty.Value `diff:"-"`
}

func (r *GithubRepository) TerraformId() string {
	return r.Id
}

func (r *GithubRepository) TerraformType() string {
	return GithubRepositoryResourceType
}

func (r *GithubRepository) CtyValue() *cty.Value {
	return r.CtyVal
}

var githubRepositoryTags = map[string]string{}

func githubRepositoryNormalizer(val *dctlcty.CtyAttributes) {
	val.SafeDelete([]string{"etag"})
	val.SafeDelete([]string{"auto_init"})
}
