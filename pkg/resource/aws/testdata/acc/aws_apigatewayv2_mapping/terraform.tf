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

resource "aws_apigatewayv2_domain_name" "example" {
    domain_name = "driftctl.example.com"

    domain_name_configuration {
        certificate_arn = aws_acm_certificate.example.arn
        endpoint_type   = "REGIONAL"
        security_policy = "TLS_1_2"
    }
}

resource "tls_private_key" "example" {
    algorithm = "RSA"
}

resource "tls_self_signed_cert" "example" {
    key_algorithm   = "RSA"
    private_key_pem = tls_private_key.example.private_key_pem

    subject {
        common_name  = "driftctl.example.com"
        organization = "ACME Examples, Inc"
    }

    validity_period_hours = 12

    allowed_uses = [
        "key_encipherment",
        "digital_signature",
        "server_auth",
    ]
}

resource "aws_acm_certificate" "example" {
    private_key      = tls_private_key.example.private_key_pem
    certificate_body = tls_self_signed_cert.example.cert_pem
}

resource "aws_apigatewayv2_api_mapping" "example" {
    api_id      = aws_apigatewayv2_api.example.id
    domain_name = aws_apigatewayv2_domain_name.example.id
    stage       = aws_apigatewayv2_stage.example.id
}
