provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.19.0"
  }
}

resource "aws_apigatewayv2_api" "example" {
  name          = "acceptance-tests"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_route" "example" {
  api_id    = aws_apigatewayv2_api.example.id
  route_key = "GET /example"
  target    = "integrations/${aws_apigatewayv2_integration.example.id}"
}

resource "aws_apigatewayv2_integration" "example" {
  api_id           = aws_apigatewayv2_api.example.id
  integration_type = "HTTP_PROXY"

  integration_method = "ANY"
  integration_uri    = "https://example.com/"
}

resource "aws_apigatewayv2_deployment" "example" {
  api_id      = aws_apigatewayv2_route.example.api_id
  description = "Example deployment"
}
