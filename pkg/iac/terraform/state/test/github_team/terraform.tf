terraform {
  required_version = ">= 0.14.4"
  required_providers {
    github = "=4.4.0"
  }
}

resource "github_team" "team1" {
  name = "team1"
  description = "test"
  privacy = "closed"
}

resource "github_team" "team2" {
  name = "team2"
  description = "test 2"
}

resource "github_team" "with_parent" {
  name = "new team with parent"
  description = "test parent team"
  parent_team_id = github_team.team1.id
  privacy = "closed"
}
