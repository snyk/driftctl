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

resource "azurerm_firewall" "example" {
    name                = "testfirewall"
    location            = data.azurerm_resource_group.example.location
    resource_group_name = data.azurerm_resource_group.example.name
}
