provider "aws" {
    region = "us-east-1"
}

terraform {
    required_providers {
        aws = "3.47.0"
    }
}

resource "aws_apigatewayv2_api" "example" {
    name          = "example-http-api"
    protocol_type = "HTTP"
}

resource "aws_apigatewayv2_stage" "example" {
    api_id = aws_apigatewayv2_api.example.id
    name   = "example-stage"
}
