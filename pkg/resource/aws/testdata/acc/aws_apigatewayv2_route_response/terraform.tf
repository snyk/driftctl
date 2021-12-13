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

resource "aws_apigatewayv2_route" "example" {
  api_id    = aws_apigatewayv2_api.example.id
  route_key = "$default"
}

resource "aws_apigatewayv2_route_response" "example" {
    api_id             = aws_apigatewayv2_api.example.id
    route_id           = aws_apigatewayv2_route.example.id
    route_response_key = "$default"
}
