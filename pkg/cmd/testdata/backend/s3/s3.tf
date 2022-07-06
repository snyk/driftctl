terraform {
    backend "s3" {
        bucket = "terraform-state-prod"
        key    = "network/terraform.tfstate"
        region = "us-east-1"
    }
}
