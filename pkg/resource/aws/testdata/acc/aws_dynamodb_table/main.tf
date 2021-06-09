provider "aws" {
  region = "us-east-1"
}

locals {
    timestamp = formatdate("YYYYMMDDhhmmss", timestamp())
}

resource "aws_dynamodb_table" "simple-dynamo-test" {
  name           = "simple-dynamo-test-${local.timestamp}"
  billing_mode   = "PROVISIONED"
  read_capacity  = 20
  write_capacity = 20
  hash_key       = "UserId"
  range_key      = "GameTitle"

  timeouts {
    create = "20m"
    delete = "30m"
  }

  attribute {
    name = "UserId"
    type = "S"
  }

  attribute {
    name = "GameTitle"
    type = "S"
  }

  attribute {
    name = "TopScore"
    type = "N"
  }

  ttl {
    attribute_name = "TimeToExist"
    enabled        = false
  }

  global_secondary_index {
    name               = "GameTitleIndex"
    hash_key           = "GameTitle"
    range_key          = "TopScore"
    write_capacity     = 10
    read_capacity      = 10
    projection_type    = "INCLUDE"
    non_key_attributes = ["UserId"]
  }

  tags = {
    Name        = "dynamodb-table-1"
    Environment = "production"
  }
}

resource "aws_dynamodb_table" "global-dynamo-test" {
  name = "global-dynamo-test-${local.timestamp}"
  hash_key = "TestTableHashKey"
  billing_mode = "PAY_PER_REQUEST"
  stream_enabled = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  timeouts {
    create = "30m"
    delete = "30m"
  }

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }
}
