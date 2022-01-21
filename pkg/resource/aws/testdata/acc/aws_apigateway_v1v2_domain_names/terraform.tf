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

    subject {
        common_name  = "example-${random_string.prefix.result}.com"
        organization = "ACME Examples, Inc"
    }
}

resource "aws_acm_certificate" "example" {
    certificate_body = tls_self_signed_cert.example.cert_pem
    private_key      = tls_private_key.example.private_key_pem
}

resource "aws_apigatewayv2_domain_name" "example" {
    domain_name = aws_acm_certificate.example.domain_name

    domain_name_configuration {
        certificate_arn = aws_acm_certificate.example.arn
        endpoint_type   = "REGIONAL"
        security_policy = "TLS_1_2"
    }
}

resource "tls_private_key" "example2" {
    algorithm = "RSA"
}

resource "tls_self_signed_cert" "example2" {
    allowed_uses = [
        "key_encipherment",
        "digital_signature",
        "server_auth",
    ]

    key_algorithm         = tls_private_key.example2.algorithm
    private_key_pem       = tls_private_key.example2.private_key_pem
    validity_period_hours = 12

    subject {
        common_name  = "example2-${random_string.prefix.result}.com"
        organization = "ACME Examples, Inc"
    }
}

resource "aws_acm_certificate" "example2" {
    certificate_body = tls_self_signed_cert.example2.cert_pem
    private_key      = tls_private_key.example2.private_key_pem
}

resource "aws_api_gateway_domain_name" "example2" {
    domain_name              = aws_acm_certificate.example2.domain_name
    regional_certificate_arn = aws_acm_certificate.example2.arn

    endpoint_configuration {
        types = ["REGIONAL"]
    }
}
