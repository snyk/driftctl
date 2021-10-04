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

resource "azurerm_route_table" "table1" {
    name                          = "table1"
    location                      = data.azurerm_resource_group.qa1.location
    resource_group_name           = data.azurerm_resource_group.qa1.name
}
