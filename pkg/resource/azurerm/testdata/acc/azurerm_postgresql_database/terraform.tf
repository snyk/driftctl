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

resource "random_password" "password" {
  length           = 16
  special          = true
  override_special = "_%@"
}

resource "azurerm_postgresql_server" "example" {
  name                = "acc-postgresql-server-${random_string.suffix.result}"
  location            = data.azurerm_resource_group.qa1.location
  resource_group_name = data.azurerm_resource_group.qa1.name

  sku_name = "B_Gen5_2"

  storage_mb                   = 5120
  backup_retention_days        = 7
  geo_redundant_backup_enabled = false
  auto_grow_enabled            = true

  administrator_login          = "psqladminun"
  administrator_login_password = random_password.password.result
  version                      = "10"
  ssl_enforcement_enabled      = true
}

resource "azurerm_postgresql_database" "example" {
  name                = "acc-test-db"
  resource_group_name = data.azurerm_resource_group.qa1.name
  server_name         = azurerm_postgresql_server.example.name
  charset             = "UTF8"
  collation           = "English_United States.1252"
}
