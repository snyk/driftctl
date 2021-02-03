provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.19.0"
  }
}

resource "aws_sqs_queue" "foo" {
  name                      = "foo"
  policy = file("policy.json")
}

resource "aws_sqs_queue" "bar" {
  name                        = "bar.fifo"
  fifo_queue                  = true
  content_based_deduplication = true
}

resource "aws_sqs_queue" "baz" {
  name   = "baz"
}

resource "aws_sqs_queue_policy" "sqs-policy" {
  queue_url = aws_sqs_queue.baz.id
  policy = <<POLICY
{
  "Id": "MYSQSPOLICY",
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Stmt1611769527792",
      "Action": ["sqs:SendMessage"],
      "Effect": "Allow",
      "Resource": "arn:aws:sqs:eu-west-3:047081014315:baz",
      "Principal": "*"
    }
  ]
}
POLICY
}
