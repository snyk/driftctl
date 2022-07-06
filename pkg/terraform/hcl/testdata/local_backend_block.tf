terraform {
    backend "local" {
        path = "terraform-state-prod/network/terraform.tfstate"
    }
}
