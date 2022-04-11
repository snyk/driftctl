provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.19.0"
  }
}

resource "aws_elasticache_cluster" "foo" {
    cluster_id           = "cluster-foo"
    engine               = "memcached"
    node_type            = "cache.t2.micro"
    num_cache_nodes      = 1
    parameter_group_name = "default.memcached1.6"
}

