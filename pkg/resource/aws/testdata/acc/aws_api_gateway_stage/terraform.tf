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

resource "aws_api_gateway_deployment" "foo" {
    rest_api_id = aws_api_gateway_rest_api.foo.id
    description = "This is foo deployment"
    stage_name = "foo"
    stage_description = "foo stage"
    variables = {
        "test" = "Test"
    }
}

resource "aws_api_gateway_stage" "foo" {
    deployment_id = aws_api_gateway_deployment.foo.id
    rest_api_id   = aws_api_gateway_rest_api.foo.id
    stage_name    = "bar"
}
