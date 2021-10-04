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

resource "azurerm_resource_group" "test-1" {
    name     = "acc-test-res-group-1"
    location = "West Europe"
}

resource "azurerm_resource_group" "test-2" {
    name     = "acc-test-res-group-2"
    location = "West Europe"
}
