terraform {
  required_version = ">= 0.14.4"
  required_providers {
    github = "=4.4.0"
  }
}

resource "github_team" "test" {
  name        = "acceptance-testing"
  description = "Acceptance tests team"
}

data "github_user" "karniwl" {
  username = "karniwl"
}

data "github_user" "chdorner-snyk" {
  username = "chdorner-snyk"
}

data "github_user" "agatakrajewska" {
  username = "agatakrajewska"
}

data "github_user" "craigfurman" {
  username = "craigfurman"
}

resource "github_team_membership" "karniwl" {
  team_id  = github_team.test.id
  username = data.github_user.karniwl.login
  role     = "maintainer"
}

resource "github_team_membership" "chdorner-snyk" {
  team_id  = github_team.test.id
  username = data.github_user.chdorner-snyk.login
  role     = "maintainer"
}

resource "github_team_membership" "agatakrajewska" {
  team_id  = github_team.test.id
  username = data.github_user.agatakrajewska.login
  role     = "maintainer"
}

resource "github_team_membership" "craigfurman" {
  team_id  = github_team.test.id
  username = data.github_user.craigfurman.login
  role     = "maintainer"
}
