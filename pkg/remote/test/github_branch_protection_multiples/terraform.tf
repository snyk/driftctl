terraform {
  required_version = ">= 0.14.4"
  required_providers {
    github = "=4.4.0"
  }
}

data "github_user" "eliecharra" {
  username = "eliecharra"
}

resource "github_repository" "repo" {
  count = 3
  name = "repo${count.index}"
  auto_init = true
}

resource "github_branch" "repo_toto" {
  count = 3
  branch = "toto"
  repository = github_repository.repo[count.index].name
  source_branch = "main"
}

resource "github_branch_protection" "main_repo" {
  count = 3
  pattern = "main"
  repository_id = github_repository.repo[count.index].node_id
  enforce_admins = true
  required_status_checks {
    strict = false
    contexts = ["ci/travis"]
  }

  required_pull_request_reviews {
    dismiss_stale_reviews = true
    dismissal_restrictions = [
      data.github_user.eliecharra.node_id
    ]
  }

  push_restrictions = [
    data.github_user.eliecharra.node_id
  ]

  allows_deletions = true
  allows_force_pushes = true
}


resource "github_branch_protection" "toto_repo" {
  count = 3
  repository_id     = github_repository.repo[count.index].node_id
  pattern         = github_branch.repo_toto[count.index].branch
  enforce_admins = true
}
