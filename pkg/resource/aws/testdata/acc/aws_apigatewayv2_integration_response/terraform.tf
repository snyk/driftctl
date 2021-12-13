provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.19.0"
  }
}

resource "aws_apigatewayv2_api" "example" {
    name                       = "example-websocket-api"
    protocol_type              = "WEBSOCKET"
    route_selection_expression = "$request.body.action"
}

resource "aws_apigatewayv2_integration" "example" {
    api_id           = aws_apigatewayv2_api.example.id
    integration_type = "MOCK"
}

resource "aws_apigatewayv2_integration_response" "example" {
    api_id                   = aws_apigatewayv2_api.example.id
    integration_id           = aws_apigatewayv2_integration.example.id
    integration_response_key = "/200/"
}
