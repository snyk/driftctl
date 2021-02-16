resource "github_repository" "public" {
  name = "public-repo"
  visibility = "public"
}

resource "github_repository" "private" {
  name = "private-repo"
  visibility = "private"
  description = "this is a private repo"
}
