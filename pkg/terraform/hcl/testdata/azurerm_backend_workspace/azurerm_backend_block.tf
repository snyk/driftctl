terraform {
    backend "azurerm" {
        resource_group_name  = "StorageAccount-ResourceGroup"
        storage_account_name = "abcd1234"
        container_name       = "states"
        key                  = "prod.terraform.tfstate"
    }
}

provider "azurerm" {
    features {}
}
