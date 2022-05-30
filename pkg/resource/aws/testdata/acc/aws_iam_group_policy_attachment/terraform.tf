provider "aws" {
    region = "us-east-1"
}

terraform {
    required_providers {
        aws = "3.19.0"
    }
}

resource "aws_iam_group" "group" {
    name = "test-acc-group"
}

resource "aws_iam_policy" "policy" {
    name        = "test-policy"
    description = "A test policy"
        policy = jsonencode({
            Version = "2012-10-17"
            Statement = [
                {
                    Action = [
                        "ec2:Describe*",
                    ]
                    Effect   = "Allow"
                    Resource = "*"
                },
            ]
        })
}

resource "aws_iam_group_policy_attachment" "test-attach" {
    group      = aws_iam_group.group.name
    policy_arn = aws_iam_policy.policy.arn
}
