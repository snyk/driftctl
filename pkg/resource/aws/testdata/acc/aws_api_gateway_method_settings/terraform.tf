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

resource "aws_api_gateway_deployment" "foo" {
    rest_api_id = aws_api_gateway_rest_api.foo.id
    triggers = {
        redeployment = sha1(jsonencode([
            aws_api_gateway_resource.foo.id,
            aws_api_gateway_method.foo.id,
            aws_api_gateway_integration.foo.id,
        ]))
    }
}

resource "aws_api_gateway_stage" "foo" {
    deployment_id = aws_api_gateway_deployment.foo.id
    rest_api_id   = aws_api_gateway_rest_api.foo.id
    stage_name    = "foo"
}

resource "aws_api_gateway_resource" "foo" {
    rest_api_id = aws_api_gateway_rest_api.foo.id
    parent_id   = aws_api_gateway_rest_api.foo.root_resource_id
    path_part   = "foo"
}

resource "aws_api_gateway_method" "foo" {
    rest_api_id   = aws_api_gateway_rest_api.foo.id
    resource_id   = aws_api_gateway_resource.foo.id
    http_method   = "GET"
    authorization = "NONE"
}

resource "aws_api_gateway_integration" "foo" {
    http_method = aws_api_gateway_method.foo.http_method
    resource_id = aws_api_gateway_resource.foo.id
    rest_api_id = aws_api_gateway_rest_api.foo.id
    type        = "MOCK"
}

resource "aws_api_gateway_method_settings" "all" {
    rest_api_id = aws_api_gateway_rest_api.foo.id
    stage_name  = aws_api_gateway_stage.foo.stage_name
    method_path = "*/*"
    settings {
        metrics_enabled = true
        logging_level   = "ERROR"
    }
}

resource "aws_api_gateway_method_settings" "path_specific" {
    rest_api_id = aws_api_gateway_rest_api.foo.id
    stage_name  = aws_api_gateway_stage.foo.stage_name
    method_path = "foo/GET"
    settings {
        metrics_enabled = true
        logging_level   = "INFO"
    }
}
