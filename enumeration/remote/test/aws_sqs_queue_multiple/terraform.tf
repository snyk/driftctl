provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "5.94.1"
  }
}

resource "aws_sqs_queue" "foo" {
  name                      = "foo"
}

resource "aws_sqs_queue" "bar" {
  name                        = "bar.fifo"
  fifo_queue                  = true
  content_based_deduplication = true
}
