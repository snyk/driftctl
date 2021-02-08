provider "aws" {
  region = "us-east-1"
}
resource "aws_sns_topic" "user_updates" {
  name = "user-updates-topic"
}

resource "aws_sns_topic" "user_updates2" {
  name = "user-updates-topic2"
}

resource "aws_sqs_queue" "user_updates_queue" {
  name = "user-updates-queue"
}

resource "aws_sns_topic_subscription" "user_updates_sqs_target" {
  filter_policy = ""
  topic_arn = aws_sns_topic.user_updates.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.user_updates_queue.arn
}

resource "aws_sns_topic_subscription" "user_updates_sqs_target2" {
  topic_arn = aws_sns_topic.user_updates2.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.user_updates_queue.arn
}