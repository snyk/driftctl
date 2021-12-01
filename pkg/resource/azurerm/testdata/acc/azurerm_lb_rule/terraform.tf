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

resource "azurerm_public_ip" "example" {
    name                = "PublicIPForLB"
    location            = "West US"
    resource_group_name = data.azurerm_resource_group.qa1.name
    allocation_method   = "Static"
}

resource "azurerm_lb" "example" {
    name                = "TestLoadBalancer"
    location            = "West US"
    resource_group_name = data.azurerm_resource_group.qa1.name

    frontend_ip_configuration {
        name                 = "PublicIPAddress"
        public_ip_address_id = azurerm_public_ip.example.id
    }
}

resource "azurerm_lb_rule" "example" {
    resource_group_name            = data.azurerm_resource_group.qa1.name
    loadbalancer_id                = azurerm_lb.example.id
    name                           = "LBRule"
    protocol                       = "Tcp"
    frontend_port                  = 3389
    backend_port                   = 3389
    frontend_ip_configuration_name = "PublicIPAddress"
}
