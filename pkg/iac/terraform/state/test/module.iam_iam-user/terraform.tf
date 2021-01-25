provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.19.0"
  }
}

module "iam_iam-user" {
  source                        = "terraform-aws-modules/iam/aws//modules/iam-user"
  version                       = "3.7.0"
  name                          = "MODULE-USER"
  create_iam_access_key         = true  # default = true
  create_iam_user_login_profile = false # default = true
}
