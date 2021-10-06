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
    name = "raphael-dev"
}

resource "azurerm_container_registry" "acr" {
    name                = "containerRegistry198745268459"
    resource_group_name = data.azurerm_resource_group.qa1.name
    location            = data.azurerm_resource_group.qa1.location
    sku                 = "Premium"
    admin_enabled       = false
    georeplications = []
}
