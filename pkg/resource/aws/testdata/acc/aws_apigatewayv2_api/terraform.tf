provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "5.94.1"
  }
}

resource "aws_apigatewayv2_api" "example" {
  name                       = "example-websocket-api"
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}
