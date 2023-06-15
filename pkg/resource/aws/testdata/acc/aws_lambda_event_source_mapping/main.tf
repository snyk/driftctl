provider "aws" {
  region = "us-east-1"
}

locals {
  timestamp = formatdate("YYYYMMDDhhmmss", timestamp())
}

resource "aws_sqs_queue" "queue1" {
  name                      = "queue1-${local.timestamp}"
  delay_seconds             = 90
  max_message_size          = 2048
  message_retention_seconds = 86400
  receive_wait_time_seconds = 10
}

resource "aws_sqs_queue" "queue2" {
  name                      = "queue2-${local.timestamp}"
  delay_seconds             = 90
  max_message_size          = 2048
  message_retention_seconds = 86400
  receive_wait_time_seconds = 10
}

resource "aws_dynamodb_table" "dynamo-event-source-mapping-test" {
  name             = "event-source-mapping-test-${local.timestamp}"
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_lambda-${local.timestamp}"

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

resource "aws_iam_policy" "policy" {
  name = "policy-${local.timestamp}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "sqs:ReceiveMessage",
        "sqs:DeleteMessage",
        "sqs:GetQueueAttributes",
        "dynamodb:GetRecords",
        "dynamodb:GetShardIterator",
        "dynamodb:DescribeStream",
        "dynamodb:ListShards",
        "dynamodb:ListStreams"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy_attachment" "policy_attachment" {
  name       = "event-source-mapping-test-attachment-${local.timestamp}"
  roles      = [aws_iam_role.iam_for_lambda.name]
  policy_arn = aws_iam_policy.policy.arn
}

resource "aws_lambda_function" "test_lambda" {
  filename      = "function.zip"
  function_name = "event-source-mapping-test-lambda-${local.timestamp}"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.test"
  runtime       = "nodejs14.x"

  environment {
    variables = {
      foo = "bar"
    }
  }
}

resource "aws_lambda_event_source_mapping" "sqs1" {
  event_source_arn = aws_sqs_queue.queue1.arn
  enabled          = true
  function_name    = aws_lambda_function.test_lambda.arn
  batch_size       = 1
}

resource "aws_lambda_event_source_mapping" "sqs2" {
  event_source_arn = aws_sqs_queue.queue2.arn
  enabled          = true
  function_name    = aws_lambda_function.test_lambda.arn
  batch_size       = 1
}

resource "aws_lambda_event_source_mapping" "dynamo" {
  event_source_arn  = aws_dynamodb_table.dynamo-event-source-mapping-test.stream_arn
  function_name     = aws_lambda_function.test_lambda.arn
  starting_position = "LATEST"
}
