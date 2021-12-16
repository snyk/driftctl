provider "aws" {
  region = "us-east-1"
}

resource "aws_route53_health_check" "http" {
  fqdn              = "moadib.net"
  port              = 80
  type              = "HTTP"
  resource_path     = "/"
  failure_threshold = "5"
  request_interval  = "30"

  tags = {
    Name = "http-moadib-net"
  }
}

resource "aws_route53_health_check" "https" {
  failure_threshold = "5"
  fqdn              = "moadib.net"
  port              = 443
  request_interval  = "30"
  resource_path     = "/"
  search_string     = "MoAdiB"
  type              = "HTTPS_STR_MATCH"
}