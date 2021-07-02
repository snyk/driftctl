terraform {
  required_version = ">= 0.14.4"
  required_providers {
    github = "=4.4.0"
  }
}

data "github_user" "wbeuil" {
  username = "wbeuil"
}

data "github_user" "driftctl" {
  username = "driftctl-acceptance-tester"
}

resource "github_team" "foo" {
  name        = "foo"
  description = "Foo team"
}

resource "github_team_membership" "foo" {
  team_id  = github_team.foo.id
  username = data.github_user.wbeuil.login
  role     = "maintainer"
}

resource "github_team_membership" "bar" {
  team_id  = github_team.foo.id
  username = data.github_user.driftctl.login
  role     = "maintainer"
}
