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

resource "aws_api_gateway_gateway_response" "foo" {
    rest_api_id   = aws_api_gateway_rest_api.foo.id
    status_code   = "401"
    response_type = "UNAUTHORIZED"
    response_templates = {
        "application/json" = "{\"message\":$context.error.messageString}"
    }
    response_parameters = {
        "gatewayresponse.header.Authorization" = "'Basic'"
    }
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
        "x-amazon-apigateway-gateway-responses": {
            "MISSING_AUTHENTICATION_TOKEN": {
                "statusCode": 403,
                "responseParameters": {
                    "gatewayresponse.header.Access-Control-Allow-Origin": "'a.b.c'",
                },
                "responseTemplates": {
                    "application/json": "{\n     \"message\": $context.error.messageString,\n     \"type\":  \"$context.error.responseType\",\n     \"stage\":  \"$context.stage\",\n     \"resourcePath\":  \"$context.resourcePath\",\n     \"stageVariables.a\":  \"$stageVariables.a\",\n     \"statusCode\": \"'403'\"\n}"
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
        "x-amazon-apigateway-gateway-responses": {
            "MISSING_AUTHENTICATION_TOKEN": {
                "statusCode": 403,
                "responseParameters": {
                    "gatewayresponse.header.Access-Control-Allow-Origin": "'a.b.c'",
                },
                "responseTemplates": {
                    "application/json": "{\n     \"message\": $context.error.messageString,\n     \"type\":  \"$context.error.responseType\",\n     \"stage\":  \"$context.stage\",\n     \"resourcePath\":  \"$context.resourcePath\",\n     \"stageVariables.a\":  \"$stageVariables.a\",\n     \"statusCode\": \"'403'\"\n}"
                }
            }
        }
    })
}
