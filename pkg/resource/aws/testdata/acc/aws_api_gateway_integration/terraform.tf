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

resource "aws_api_gateway_rest_api" "baz" {
    name        = "baz"
    description = "This is baz API"
    body = jsonencode({
        swagger = "2.0"
        info = {
            title   = "test"
            version = "2017-04-20T04:08:08Z"
        }
        schemes = ["https"]
        paths = {
            "/test" = {
                get = {
                    responses = {
                        "200" = {
                            description = "OK"
                        }
                    }
                    x-amazon-apigateway-integration = {
                        httpMethod = "GET"
                        type       = "HTTP"
                        responses = {
                            default = {
                                statusCode = 200
                            }
                        }
                        uri = "https://aws.amazon.com/"
                    }
                }
            }
        }
    })
}
