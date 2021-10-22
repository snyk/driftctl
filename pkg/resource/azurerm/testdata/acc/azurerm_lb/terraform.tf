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

resource "random_uuid" "rgId" {}

resource "azurerm_resource_group" "default" {
    name     = "AccLoadBalancerRG-${random_uuid.rgId.result}"
    location = "West Europe"
}

resource "azurerm_public_ip" "default" {
    name                = "AccPublicIPForLB"
    location            = "West US"
    resource_group_name = azurerm_resource_group.default.name
    allocation_method   = "Static"
}

resource "azurerm_lb" "default" {
    name                = "AccTestLoadBalancer-7128046527"
    location            = "West US"
    resource_group_name = azurerm_resource_group.default.name

    frontend_ip_configuration {
        name                 = "PublicIPAddress"
        public_ip_address_id = azurerm_public_ip.default.id
    }
}
