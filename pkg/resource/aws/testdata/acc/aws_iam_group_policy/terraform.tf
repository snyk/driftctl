provider "aws" {
    region = "us-east-1"
}

terraform {
    required_providers {
        aws = "3.19.0"
    }
}

resource "aws_iam_group" "my_developers" {
    name = "developers"
    path = "/users/"
}

resource "aws_iam_group_policy" "my_developer_policy" {
    name  = "my_developer_policy"
    group = aws_iam_group.my_developers.name
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

