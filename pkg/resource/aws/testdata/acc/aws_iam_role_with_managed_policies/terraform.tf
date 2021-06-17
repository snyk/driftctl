provider "aws" {
    region = "us-east-1"
}

terraform {
    required_providers {
        aws = {
            version = "3.45.0"
        }
    }
}

resource "aws_iam_role" "b" {
    name = "test_role"

    managed_policy_arns = [aws_iam_policy.b.arn]

    # Terraform's "jsonencode" function converts a
    # Terraform expression result to valid JSON syntax.
    assume_role_policy = jsonencode({
        Version = "2012-10-17"
        Statement = [
            {
                Action = "sts:AssumeRole"
                Effect = "Allow"
                Sid    = ""
                Principal = {
                    Service = "ec2.amazonaws.com"
                }
            },
        ]
    })
}


resource "aws_iam_policy" "b" {
    name        = "b"
    path        = "/"
    description = "bbb"

    # Terraform's "jsonencode" function converts a
    # Terraform expression result to valid JSON syntax.
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
