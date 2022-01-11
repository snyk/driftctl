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
    protocol_type = "WEBSOCKET"
}

resource "aws_apigatewayv2_model" "example" {
    api_id       = aws_apigatewayv2_api.example.id
    content_type = "application/json"
    name         = "example"

    schema = <<EOF
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "title": "ExampleModel",
  "type": "object",
  "properties": {
    "id": { "type": "string" }
  }
}
EOF
}
