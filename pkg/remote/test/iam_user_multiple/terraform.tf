
provider "aws" {
  version = "3.5.0"
  region  = "eu-west-3"
}

resource "aws_iam_user" "testuser" {
  count         = 3
  name          = "test-driftctl-${count.index}"
  force_destroy = true
  path          = "/test/"
  tags = {
    foo = "bar"
  }
}