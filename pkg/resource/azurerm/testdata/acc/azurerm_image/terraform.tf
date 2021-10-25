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

resource "azurerm_managed_disk" "example" {
  name                 = "acctestmd"
  location             = data.azurerm_resource_group.default.location
  resource_group_name  = data.azurerm_resource_group.default.name
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = "25"

  tags = {
    environment = "acc"
  }
}

resource "azurerm_image" "example2" {
  name                = "acctest2"
  location            = data.azurerm_resource_group.default.location
  resource_group_name = data.azurerm_resource_group.default.name

  os_disk {
    os_type         = "Linux"
    os_state        = "Generalized"
    size_gb         = 30
    managed_disk_id = azurerm_managed_disk.example.id
  }
}
