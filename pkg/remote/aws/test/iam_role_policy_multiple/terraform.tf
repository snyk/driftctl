provider "aws" {
  version = "3.5.0"
  region  = "eu-west-3"
}

resource "aws_iam_role" "test_role" {
  name = "test_role_${count.index}"
  count = 2
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
}

resource "aws_iam_role_policy" "test_policy_role0" {
  count = 3
  name = "policy-role0-${count.index}"
  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": "account:*",
            "Resource": "*"
        }
    ]
}
EOF
  role = aws_iam_role.test_role[0].id
}

resource "aws_iam_role_policy" "test_policy_role1" {
  count = 3
  name = "policy-role1-${count.index}"
  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": "account:*",
            "Resource": "*"
        }
    ]
}
EOF
  role = aws_iam_role.test_role[1].id
}