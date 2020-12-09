provider "aws" {
  region = "eu-west-3"
  version = "3.19.0"
}

resource "aws_signer_signing_profile" "example" {
  name_prefix = "example"
  platform_id = "AWSLambda-SHA384-ECDSA"
}
resource "aws_lambda_code_signing_config" "example" {
  allowed_publishers {
    signing_profile_version_arns = [aws_signer_signing_profile.example.version_arn]
  }
  policies {
    untrusted_artifact_on_deployment = "Enforce"
  }
}

resource "aws_iam_role" "test-iam_for_lambda" {
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
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_lambda_function" "foo" {
  filename      = "lambda.zip"
  function_name = "foo"
  handler       = "lambda.handler"
  runtime       = "nodejs12.x"
  role          = aws_iam_role.test-iam_for_lambda.arn
  code_signing_config_arn = aws_lambda_code_signing_config.example.arn
}