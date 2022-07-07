provider "aws" {
  region = "eu-west-1"
}
provider "aws" {
  alias  = "eu-west-3"
  region = "eu-west-3"
}
provider "aws" {
  alias  = "ap-northeast-1"
  region = "ap-northeast-1"
}
resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_lambda"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_lambda_function" "func1" {
  filename      = "function.zip"
  function_name = "example_lambda_name1"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "go1.x"
}
resource "aws_lambda_function" "func2" {
  filename      = "function.zip"
  function_name = "example_lambda_name2"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "go1.x"
}

resource "aws_lambda_function" "func1w3" {
  provider = aws.eu-west-3
  filename      = "function.zip"
  function_name = "example_lambda_name1"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "go1.x"
}
resource "aws_lambda_function" "func2w3" {
  provider = aws.eu-west-3
  filename      = "function.zip"
  function_name = "example_lambda_name2"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "go1.x"
}

resource "aws_lambda_function" "func1ap" {
  provider = aws.ap-northeast-1
  filename      = "function.zip"
  function_name = "example_lambda_name1"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "go1.x"
}
resource "aws_lambda_function" "func2ap" {
  provider = aws.ap-northeast-1
  filename      = "function.zip"
  function_name = "example_lambda_name2"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "go1.x"
}

// <editor-fold desc="bucket-martin-test-drift">
resource "aws_s3_bucket" "bucket" {
  bucket = "bucket-martin-test-drift"
}


resource "aws_lambda_permission" "allow1_bucket" {
  statement_id  = "Allow1ExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.func1.arn
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.bucket.arn
}

resource "aws_lambda_permission" "allow2_bucket" {
  statement_id  = "Allow2ExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.func2.arn
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.bucket.arn
}
resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = aws_s3_bucket.bucket.id

  lambda_function {
    lambda_function_arn = aws_lambda_function.func1.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "AWSLogs/"
    filter_suffix       = ".log"
  }

  lambda_function {
    lambda_function_arn = aws_lambda_function.func2.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "OtherLogs/"
    filter_suffix       = ".log"
  }
}

resource "aws_s3_bucket_policy" "bucket_policy" {
  bucket = aws_s3_bucket.bucket.id
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "PublicReadGetObject",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::bucket-martin-test-drift/*"
    }
  ]
}
POLICY
}
resource "aws_s3_bucket_analytics_configuration" "analytics-cls-test" {
  bucket = aws_s3_bucket.bucket.bucket
  name   = "Analytics_Bucket1"
  storage_class_analysis {
    data_export {
      destination {
        s3_bucket_destination {
          bucket_arn = aws_s3_bucket.bucket.arn
        }
      }
    }
  }
}

resource "aws_s3_bucket_analytics_configuration" "analytics-cls-test2" {
  bucket = aws_s3_bucket.bucket.bucket
  name   = "Analytics2_Bucket1"
  storage_class_analysis {
    data_export {
      destination {
        s3_bucket_destination {
          bucket_arn = aws_s3_bucket.bucket.arn
        }
      }
    }
  }
}

resource "aws_s3_bucket_inventory" "inventory-cls-test" {
  bucket = aws_s3_bucket.bucket.id
  name   = "Inventory_Bucket1"
  included_object_versions = "All"
  schedule {
    frequency = "Daily"
  }
  destination {
    bucket {
      format     = "ORC"
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}

resource "aws_s3_bucket_inventory" "inventory-cls-test2" {
  bucket = aws_s3_bucket.bucket.id
  name   = "Inventory2_Bucket1"
  included_object_versions = "All"
  schedule {
    frequency = "Daily"
  }
  destination {
    bucket {
      format     = "ORC"
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}

resource "aws_s3_bucket_metric" "metrics-cls-test" {
  bucket = aws_s3_bucket.bucket.id
  name   = "Metrics_Bucket1"
}

resource "aws_s3_bucket_metric" "metrics-cls-test2" {
  bucket =aws_s3_bucket.bucket.id
  name   = "Metrics2_Bucket1"
}
// </editor-fold>

// <editor-fold desc="bucket-martin-test-drift2">
resource "aws_s3_bucket" "bucket2" {
  provider = aws.eu-west-3
  bucket = "bucket-martin-test-drift2"
}

resource "aws_lambda_permission" "allow1_bucket2" {
  provider = aws.eu-west-3
  statement_id  = "Allow1ExecutionFromS3Bucket2"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.func1w3.arn
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.bucket2.arn
}

resource "aws_lambda_permission" "allow2_bucket2" {
  provider = aws.eu-west-3
  statement_id  = "Allow2ExecutionFromS3Bucket2"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.func2w3.arn
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.bucket2.arn
}
resource "aws_s3_bucket_notification" "bucket_notification2" {
  provider = aws.eu-west-3
  bucket = aws_s3_bucket.bucket2.id

  lambda_function {
    lambda_function_arn = aws_lambda_function.func1w3.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "AWSLogs/"
    filter_suffix       = ".log"
  }

  lambda_function {
    lambda_function_arn = aws_lambda_function.func2w3.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "OtherLogs/"
    filter_suffix       = ".log"
  }
}

resource "aws_s3_bucket_policy" "bucket2_policy" {
  provider = aws.eu-west-3
  bucket = aws_s3_bucket.bucket2.id
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "PublicReadGetObject",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::bucket-martin-test-drift2/*"
    }
  ]
}
POLICY
}
resource "aws_s3_bucket_analytics_configuration" "analytics-cls-testbucket2" {
  provider = aws.eu-west-3
  bucket = aws_s3_bucket.bucket2.bucket
  name   = "Analytics_Bucket2"
  storage_class_analysis {
    data_export {
      destination {
        s3_bucket_destination {
          bucket_arn = aws_s3_bucket.bucket2.arn
        }
      }
    }
  }
}

resource "aws_s3_bucket_analytics_configuration" "analytics-cls-test2bucket2" {
  provider = aws.eu-west-3
  bucket = aws_s3_bucket.bucket2.bucket
  name   = "Analytics2_Bucket2"
  storage_class_analysis {
    data_export {
      destination {
        s3_bucket_destination {
          bucket_arn = aws_s3_bucket.bucket2.arn
        }
      }
    }
  }
}

resource "aws_s3_bucket_inventory" "inventory-cls-testbucket2" {
  provider = aws.eu-west-3
  bucket = aws_s3_bucket.bucket2.id
  name   = "Inventory_Bucket2"
  included_object_versions = "All"
  schedule {
    frequency = "Daily"
  }
  destination {
    bucket {
      format     = "ORC"
      bucket_arn = aws_s3_bucket.bucket2.arn
    }
  }
}

resource "aws_s3_bucket_inventory" "inventory-cls-test2bucket2" {
  provider = aws.eu-west-3
  bucket = aws_s3_bucket.bucket2.id
  name   = "Inventory2_Bucket2"
  included_object_versions = "All"
  schedule {
    frequency = "Daily"
  }
  destination {
    bucket {
      format     = "ORC"
      bucket_arn = aws_s3_bucket.bucket2.arn
    }
  }
}

resource "aws_s3_bucket_metric" "metrics-cls-testbucket2" {
  provider = aws.eu-west-3
  bucket = aws_s3_bucket.bucket2.id
  name   = "Metrics_Bucket2"
}

resource "aws_s3_bucket_metric" "metrics-cls-test2bucket2" {
  provider = aws.eu-west-3
  bucket =aws_s3_bucket.bucket2.id
  name   = "Metrics2_Bucket2"
}
// </editor-fold>

// <editor-fold desc="bucket-martin-test-drift3">
resource "aws_s3_bucket" "bucket3" {
  provider = aws.ap-northeast-1
  bucket = "bucket-martin-test-drift3"
}

resource "aws_lambda_permission" "allow1_bucket3" {
  provider = aws.ap-northeast-1
  statement_id  = "Allow1ExecutionFromS3Bucket3"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.func1ap.arn
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.bucket3.arn
}

resource "aws_lambda_permission" "allow2_bucket3" {
  provider = aws.ap-northeast-1
  statement_id  = "Allow2ExecutionFromS3Bucket3"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.func2ap.arn
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.bucket3.arn
}
resource "aws_s3_bucket_notification" "bucket_notification3" {
  provider = aws.ap-northeast-1
  bucket = aws_s3_bucket.bucket3.id

  lambda_function {
    lambda_function_arn = aws_lambda_function.func1ap.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "AWSLogs/"
    filter_suffix       = ".log"
  }

  lambda_function {
    lambda_function_arn = aws_lambda_function.func2ap.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "OtherLogs/"
    filter_suffix       = ".log"
  }
}

resource "aws_s3_bucket_policy" "bucket3_policy" {
  provider = aws.ap-northeast-1
  bucket = aws_s3_bucket.bucket3.id
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "PublicReadGetObject",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::bucket-martin-test-drift3/*"
    }
  ]
}
POLICY
}
resource "aws_s3_bucket_analytics_configuration" "analytics-cls-testbucket3" {
  provider = aws.ap-northeast-1
  bucket = aws_s3_bucket.bucket3.bucket
  name   = "Analytics_Bucket3"
  storage_class_analysis {
    data_export {
      destination {
        s3_bucket_destination {
          bucket_arn = aws_s3_bucket.bucket3.arn
        }
      }
    }
  }
}

resource "aws_s3_bucket_analytics_configuration" "analytics-cls-test2bucket3" {
  provider = aws.ap-northeast-1
  bucket = aws_s3_bucket.bucket3.bucket
  name   = "Analytics2_Bucket3"
  storage_class_analysis {
    data_export {
      destination {
        s3_bucket_destination {
          bucket_arn = aws_s3_bucket.bucket3.arn
        }
      }
    }
  }
}

resource "aws_s3_bucket_inventory" "inventory-cls-testbucket3" {
  provider = aws.ap-northeast-1
  bucket = aws_s3_bucket.bucket3.id
  name   = "Inventory_Bucket3"
  included_object_versions = "All"
  schedule {
    frequency = "Daily"
  }
  destination {
    bucket {
      format     = "ORC"
      bucket_arn = aws_s3_bucket.bucket3.arn
    }
  }
}

resource "aws_s3_bucket_inventory" "inventory-cls-test2bucket3" {
  provider = aws.ap-northeast-1
  bucket = aws_s3_bucket.bucket3.id
  name   = "Inventory2_Bucket3"
  included_object_versions = "All"
  schedule {
    frequency = "Daily"
  }
  destination {
    bucket {
      format     = "ORC"
      bucket_arn = aws_s3_bucket.bucket3.arn
    }
  }
}

resource "aws_s3_bucket_metric" "metrics-cls-testbucket3" {
  provider = aws.ap-northeast-1
  bucket = aws_s3_bucket.bucket3.id
  name   = "Metrics_Bucket3"
}

resource "aws_s3_bucket_metric" "metrics-cls-test2bucket3" {
  provider = aws.ap-northeast-1
  bucket =aws_s3_bucket.bucket3.id
  name   = "Metrics2_Bucket3"
}
// </editor-fold>