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

resource "azurerm_storage_account" "example" {
    name                     = "testaccdriftctl"
    resource_group_name      = data.azurerm_resource_group.qa1.name
    location                 = data.azurerm_resource_group.qa1.location
    account_tier             = "Standard"
    account_replication_type = "GRS"
    allow_blob_public_access = true
    tags = {
        environment = "dev"
    }
}

resource "azurerm_storage_account" "noblob" {
    name                     = "testaccdriftctlnoblob"
    resource_group_name      = data.azurerm_resource_group.qa1.name
    location                 = data.azurerm_resource_group.qa1.location
    account_tier             = "Premium"
    account_replication_type = "LRS"
    account_kind = "FileStorage"
}

resource "azurerm_storage_container" "private" {
    name                  = "private"
    storage_account_name  = azurerm_storage_account.example.name
    container_access_type = "private"
}

resource "azurerm_storage_container" "container" {
    name                  = "container"
    storage_account_name  = azurerm_storage_account.example.name
    container_access_type = "container"
}

resource "azurerm_storage_container" "blob" {
    name                  = "blob"
    storage_account_name  = azurerm_storage_account.example.name
    container_access_type = "blob"
}
