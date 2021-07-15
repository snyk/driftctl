provider "aws" {
  region = "eu-west-1"
}

// <editor-fold desc="test">
resource "aws_iam_policy" "test_ro" {
  name = "test"
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

resource "aws_iam_user_policy_attachment" "test-attach" {
  user       = aws_iam_user.test.name
  policy_arn = aws_iam_policy.test_ro.arn
}

resource "aws_iam_policy" "test_ro2" {
  name = "test2"

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

resource "aws_iam_user_policy_attachment" "test-attach2" {
  user       = aws_iam_user.test.name
  policy_arn = aws_iam_policy.test_ro2.arn
}

resource "aws_iam_policy" "test_ro3" {
  name = "test3"

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

resource "aws_iam_user_policy_attachment" "test-attach3" {
  user       = aws_iam_user.test.name
  policy_arn = aws_iam_policy.test_ro3.arn
}

resource "aws_iam_policy" "test_ro4" {
  name = "test4"
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

resource "aws_iam_user_policy_attachment" "test-attach21" {
  user       = aws_iam_user.test2.name
  policy_arn = aws_iam_policy.test_ro.arn
}

resource "aws_iam_user_policy_attachment" "test-attach22" {
  user       = aws_iam_user.test2.name
  policy_arn = aws_iam_policy.test_ro2.arn
}

resource "aws_iam_user_policy_attachment" "test-attach23" {
  user       = aws_iam_user.test2.name
  policy_arn = aws_iam_policy.test_ro3.arn
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

resource "aws_iam_user_policy_attachment" "test-attach31" {
  user       = aws_iam_user.test3.name
  policy_arn = aws_iam_policy.test_ro.arn
}

resource "aws_iam_user_policy_attachment" "test-attach32" {
  user       = aws_iam_user.test3.name
  policy_arn = aws_iam_policy.test_ro2.arn
}

resource "aws_iam_user_policy_attachment" "test-attach33" {
  user       = aws_iam_user.test3.name
  policy_arn = aws_iam_policy.test_ro3.arn
}


resource "aws_iam_user" "test3" {
  name = "loadbalancer3"
  path = "/system/"
}

resource "aws_iam_access_key" "test3" {
  user = aws_iam_user.test3.name
}
// </editor-fold>

resource "aws_iam_policy_attachment" "test-attach4" {
  name       = "test-attachment"
  users      = [aws_iam_user.test.name, aws_iam_user.test2.name, aws_iam_user.test3.name]
  policy_arn = aws_iam_policy.test_ro4.arn
}