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

resource "random_string" "suffix" {
  length  = 12
  upper   = false
  special = false
}

resource "azurerm_container_registry" "acr" {
  name                = "containerRegistryAcc${random_string.suffix.result}"
  resource_group_name = data.azurerm_resource_group.qa1.name
  location            = data.azurerm_resource_group.qa1.location
  sku                 = "Premium"
  admin_enabled       = false
  georeplications     = []
}
