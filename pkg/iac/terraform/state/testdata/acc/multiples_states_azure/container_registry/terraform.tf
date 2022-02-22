terraform {
    required_providers {
        azurerm = {
            source  = "hashicorp/azurerm"
            version = "~> 2.71.0"
        }
    }
    backend "azurerm" {
        // WARNING: If you change the resource group you also have to change it the golang unit test file
        resource_group_name  = "driftctl-qa-1"
        storage_account_name = "driftctlacctest"
        container_name       = "foobar"
        key                  = "states/valid/registry/terraform.tfstate"
    }
}

provider "azurerm" {
    features {}
}

data "azurerm_resource_group" "qa1" {
    // WARNING: If you change the resource group you also have to change it the golang unit test file
    name = "driftctl-qa-1"
}

resource "random_string" "suffix" {
    length  = 12
    upper   = false
    special = false
}

resource "azurerm_container_registry" "registry" {
    name                = "dctltestmultiplestate${random_string.suffix.result}"
    resource_group_name = data.azurerm_resource_group.qa1.name
    location            = data.azurerm_resource_group.qa1.location
    sku                 = "Premium"
    admin_enabled       = false
    georeplications     = []
}

