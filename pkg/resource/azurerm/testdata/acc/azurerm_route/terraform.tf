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

resource "azurerm_route" "route1" {
    count = 2
    name                = "route${count.index}"
    resource_group_name = data.azurerm_resource_group.qa1.name
    route_table_name    = azurerm_route_table.table1.name
    address_prefix      = "10.${count.index+1}.0.0/24"
    next_hop_type       = "vnetlocal"
}

resource "azurerm_route_table" "table2" {
    name                          = "table2"
    location                      = data.azurerm_resource_group.qa1.location
    resource_group_name           = data.azurerm_resource_group.qa1.name
}

resource "azurerm_route" "route2" {
    count = 2
    name                = "route${count.index}"
    resource_group_name = data.azurerm_resource_group.qa1.name
    route_table_name    = azurerm_route_table.table2.name
    address_prefix      = "10.${count.index+3}.0.0/24"
    next_hop_type       = "vnetlocal"
}
