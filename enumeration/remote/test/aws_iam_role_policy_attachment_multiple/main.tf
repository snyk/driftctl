provider "aws" {
  region = "eu-west-1"
}

// <editor-fold desc="test">
resource "aws_iam_role" "test" {
  name = "test-role"

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

resource "aws_iam_policy" "policy" {
  name        = "test-policy"
  description = "A test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test-attach1" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.policy.arn
}

resource "aws_iam_policy" "policy2" {
  name        = "test-policy2"
  description = "A test policy 2"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test-attach2" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.policy2.arn
}
// </editor-fold>

// <editor-fold desc="test2">

resource "aws_iam_role" "test2" {
  name = "test-role2"

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

resource "aws_iam_role_policy_attachment" "test-attach3" {
  role       = aws_iam_role.test2.name
  policy_arn = aws_iam_policy.policy.arn
}

resource "aws_iam_role_policy_attachment" "test-attach4" {
  role       = aws_iam_role.test2.name
  policy_arn = aws_iam_policy.policy2.arn
}

// </editor-fold>

resource "aws_iam_policy" "policy3" {
  name        = "test-policy3"
  description = "A test policy 3"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy_attachment" "test-attach5" {
  name       = "test-attachment5"
  roles = [aws_iam_role.test.name, aws_iam_role.test2.name]
  policy_arn = aws_iam_policy.policy3.arn
}