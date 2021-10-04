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

    route {
        address_prefix = "10.1.0.0/24"
        name           = "route1"
        next_hop_type  = "vnetlocal"
    }

    route {
        address_prefix = "10.2.0.0/24"
        name           = "route2"
        next_hop_type  = "vnetlocal"
    }
}

resource "azurerm_route_table" "table2" {
    name                          = "table2"
    location                      = data.azurerm_resource_group.qa1.location
    resource_group_name           = data.azurerm_resource_group.qa1.name

    route {
        address_prefix = "10.3.0.0/24"
        name           = "route3"
        next_hop_type  = "vnetlocal"
    }

    route {
        address_prefix = "10.4.0.0/24"
        name           = "route4"
        next_hop_type  = "vnetlocal"
    }
}
