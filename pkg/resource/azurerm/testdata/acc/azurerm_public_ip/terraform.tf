terraform {
    required_providers {
        azurerm = {
            source  = "hashicorp/azurerm"
            version = "~> 2.71.0"
        }
    }
}

provider "azurerm" {
    features {}
}

data "azurerm_resource_group" "example" {
   name = "driftctl-qa-1"
}

resource "azurerm_public_ip" "example" {
    name                = "ip1"
    resource_group_name = data.azurerm_resource_group.example.name
    location            = data.azurerm_resource_group.example.location
    allocation_method   = "Static"
}
