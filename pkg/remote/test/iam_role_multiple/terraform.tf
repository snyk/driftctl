provider "aws" {
  version = "3.5.0"
  region  = "eu-west-3"
}

resource "aws_iam_role" "test_role" {
  name = "test_role_${count.index}"
  count = 3
  max_session_duration = 3600
  path = "/test/"
  force_detach_policies = true
assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  tags = {
    foo = "bar${count.index}"
  }
}