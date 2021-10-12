provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.19.0"
  }
}

resource "random_string" "prefix" {
    length  = 6
    upper   = false
    special = false
}

resource "tls_private_key" "example" {
    algorithm = "RSA"
}

resource "tls_self_signed_cert" "example" {
    allowed_uses = [
        "key_encipherment",
        "digital_signature",
        "server_auth",
    ]

    key_algorithm         = tls_private_key.example.algorithm
    private_key_pem       = tls_private_key.example.private_key_pem
    validity_period_hours = 12

    dns_names = ["example-${random_string.prefix.result}.com"]

    subject {
        common_name  = "example-${random_string.prefix.result}.com"
        organization = "ACME Examples, Inc"
    }
}

resource "aws_acm_certificate" "example" {
    certificate_body = tls_self_signed_cert.example.cert_pem
    private_key      = tls_private_key.example.private_key_pem
}

resource "aws_api_gateway_domain_name" "foo" {
    domain_name              = aws_acm_certificate.example.domain_name
    regional_certificate_arn = aws_acm_certificate.example.arn

    endpoint_configuration {
        types = ["REGIONAL"]
    }
}

resource "aws_api_gateway_rest_api" "foo" {
    name        = "foo"
    description = "This is foo API"
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

resource "aws_api_gateway_deployment" "foo" {
    rest_api_id = aws_api_gateway_rest_api.foo.id
}

resource "aws_api_gateway_stage" "foo" {
    deployment_id = aws_api_gateway_deployment.foo.id
    rest_api_id   = aws_api_gateway_rest_api.foo.id
    stage_name    = "foo"
}

resource "aws_api_gateway_base_path_mapping" "foo" {
    api_id      = aws_api_gateway_rest_api.foo.id
    stage_name  = aws_api_gateway_stage.foo.stage_name
    domain_name = aws_api_gateway_domain_name.foo.domain_name
    base_path = "foo"
}

resource "aws_api_gateway_base_path_mapping" "bar" {
    api_id      = aws_api_gateway_rest_api.foo.id
    stage_name  = aws_api_gateway_stage.foo.stage_name
    domain_name = aws_api_gateway_domain_name.foo.domain_name
    base_path = "bar"
}

resource "aws_api_gateway_base_path_mapping" "empty" {
    api_id      = aws_api_gateway_rest_api.foo.id
    stage_name  = aws_api_gateway_stage.foo.stage_name
    domain_name = aws_api_gateway_domain_name.foo.domain_name
}
