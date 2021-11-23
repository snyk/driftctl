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

data "azurerm_resource_group" "example" {
    name = "driftctl-qa-1"
}

resource "azurerm_private_dns_zone" "testzone" {
    name                = "this-zone-is-a-test-for-driftctl.com"
    resource_group_name = data.azurerm_resource_group.example.name
}

resource "azurerm_private_dns_txt_record" "othertesttxt" {
    name                = "othertesttxt"
    zone_name           = azurerm_private_dns_zone.testzone.name
    resource_group_name = data.azurerm_resource_group.example.name
    ttl                 = 300
    record {
        value = "this is value line 1"
    }

    record {
        value = "this is value line 2"
    }
}

resource "azurerm_private_dns_txt_record" "testtxt" {
    name                = "testtxt"
    zone_name           = azurerm_private_dns_zone.testzone.name
    resource_group_name = data.azurerm_resource_group.example.name
    ttl                 = 300
    record {
        value = "this is value line 3"
    }
}

