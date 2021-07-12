provider "aws" {
  region = "eu-west-1"
}

// <editor-fold desc="test">
resource "aws_iam_user_policy" "test_ro" {
  name = "test"
  user = aws_iam_user.test.name

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

resource "aws_iam_user_policy" "test_ro2" {
  name = "test2"
  user = aws_iam_user.test.name

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

resource "aws_iam_user_policy" "test_ro3" {
  name = "test3"
  user = aws_iam_user.test.name

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

resource "aws_iam_user_policy" "test_ro4" {
  name = "test4"
  user = aws_iam_user.test.name

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

resource "aws_iam_user" "test" {
  name = "loadbalancer"
  path = "/system/"
}

resource "aws_iam_access_key" "test" {
  user = aws_iam_user.test.name
}
// </editor-fold>

// <editor-fold desc="test2">
resource "aws_iam_user_policy" "test2_ro" {
  name = "test2"
  user = aws_iam_user.test2.name

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

resource "aws_iam_user_policy" "test2_ro2" {
  name = "test22"
  user = aws_iam_user.test2.name

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

resource "aws_iam_user_policy" "test2_ro3" {
  name = "test23"
  user = aws_iam_user.test2.name

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

resource "aws_iam_user_policy" "test2_ro4" {
  name = "test24"
  user = aws_iam_user.test2.name

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

resource "aws_iam_user" "test2" {
  name = "loadbalancer2"
  path = "/system/"
}

resource "aws_iam_access_key" "test2" {
  user = aws_iam_user.test2.name
}
// </editor-fold>

// <editor-fold desc="test3">
resource "aws_iam_user_policy" "test3_ro" {
  name = "test3"
  user = aws_iam_user.test3.name

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

resource "aws_iam_user_policy" "test3_ro2" {
  name = "test32"
  user = aws_iam_user.test3.name

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

resource "aws_iam_user_policy" "test3_ro3" {
  name = "test33"
  user = aws_iam_user.test3.name

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

resource "aws_iam_user_policy" "test3_ro4" {
  name = "test34"
  user = aws_iam_user.test3.name

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

resource "aws_iam_user" "test3" {
  name = "loadbalancer3"
  path = "/system/"
}

resource "aws_iam_access_key" "test3" {
  user = aws_iam_user.test3.name
}
// </editor-fold>

