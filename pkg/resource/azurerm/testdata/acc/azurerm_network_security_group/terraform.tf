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

data "azurerm_resource_group" "default" {
  name = "driftctl-qa-1"
}

resource "random_string" "suffix" {
  length  = 12
  upper   = false
  special = false
}

resource "azurerm_network_security_group" "test" {
  name                = "acceptanceTestSecurityGroup-${random_string.suffix.result}"
  location            = data.azurerm_resource_group.default.location
  resource_group_name = data.azurerm_resource_group.default.name

  security_rule {
    name                       = "test123"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
}
