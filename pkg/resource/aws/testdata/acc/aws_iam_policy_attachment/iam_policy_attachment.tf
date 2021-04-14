resource "aws_iam_role" "yaddi" {
    name               = "driftctl-lambda-role2"
    path               = "/service-role/"
    assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "yadda" {
    name = "policy"

    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "sqs:ReceiveMessage",
        "sqs:DeleteMessage",
        "sqs:GetQueueAttributes"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy_attachment" "yaddattach" {
    name = "yaddattachment2"
    roles = [aws_iam_role.yaddi.name]
    policy_arn = aws_iam_policy.yadda.arn
}
