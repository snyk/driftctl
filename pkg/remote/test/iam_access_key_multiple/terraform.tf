
provider "aws" {
  version = "3.5.0"
  region  = "eu-west-3"
}

resource "aws_iam_user" "testuser" {
  name = "test-driftctl"
}

resource "aws_iam_access_key" "key" {
  count = 2
  user  = aws_iam_user.testuser.name
}

resource "aws_iam_user" "testuser2" {
  name = "test-driftctl2"
}

resource "aws_iam_access_key" "key2" {
  count = 2
  user  = aws_iam_user.testuser2.name
}