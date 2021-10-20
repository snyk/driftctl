provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.19.0"
  }
}

resource "aws_api_gateway_rest_api" "foo" {
    name        = "foo"
    description = "This is foo API"
}

resource "aws_api_gateway_model" "foo" {
    rest_api_id  = aws_api_gateway_rest_api.foo.id
    name         = "foo"
    description  = "a JSON schema"
    content_type = "application/json"
    schema = <<EOF
{
  "type": "object"
}
EOF
}

resource "aws_api_gateway_rest_api" "bar" {
    name        = "bar"
    description = "This is bar API"
    body = jsonencode({
        openapi = "3.0.1"
        info = {
            title   = "example"
            version = "1.0"
        }
        paths = {
            "/path1" = {
                get = {
                    x-amazon-apigateway-integration = {
                        httpMethod           = "GET"
                        payloadFormatVersion = "1.0"
                        type                 = "HTTP_PROXY"
                        uri                  = "https://ip-ranges.amazonaws.com/ip-ranges.json"
                    }
                }
            }
        }
    })
}

resource "aws_api_gateway_model" "bar" {
    rest_api_id  = aws_api_gateway_rest_api.bar.id
    name         = "bar"
    description  = "a JSON schema"
    content_type = "application/json"
    schema = <<EOF
{
  "type": "object"
}
EOF
}
