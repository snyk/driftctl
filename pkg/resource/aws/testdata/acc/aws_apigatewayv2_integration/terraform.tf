provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.19.0"
  }
}

resource "aws_apigatewayv2_api" "test" {
    name          = "test"
    protocol_type = "HTTP"
    body = jsonencode({
        openapi = "3.0.1"
        paths = {
            "/path1" = {
                get = {
                    "x-amazon-apigateway-integration": {
                        "payloadFormatVersion": "1.0",
                        "type": "HTTP_PROXY",
                        "httpMethod": "GET",
                        "uri": "https://example.com",
                        "connectionType": "INTERNET"
                    },
                    "responses" : {
                        "default" : {
                            "description" : "Default response for GET /path1"
                        }
                    },
                }
            }
        }
    })
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
