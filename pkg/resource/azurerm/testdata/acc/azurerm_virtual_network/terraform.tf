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

data "azurerm_resource_group" "qa1" {
    name = "driftctl-qa-1"
}

resource "azurerm_virtual_network" "test" {
    name                = "network1"
    location            = data.azurerm_resource_group.qa1.location
    resource_group_name = data.azurerm_resource_group.qa1.name
    address_space       = ["10.0.0.0/16"]
    dns_servers         = ["10.0.0.4", "10.0.0.5"]
}
