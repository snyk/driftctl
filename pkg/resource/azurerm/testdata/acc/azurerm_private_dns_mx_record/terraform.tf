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

resource "azurerm_private_dns_mx_record" "othertestmx" {
    name                = "othertestmx"
    zone_name           = azurerm_private_dns_zone.testzone.name
    resource_group_name = data.azurerm_resource_group.example.name
    ttl                 = 300
    record {
        preference = 10
        exchange   = "mx.thisisatestusingtf.com"
    }

    record {
        preference = 20
        exchange   = "backupmx.thisisatestusingtf.com"
    }
}

resource "azurerm_private_dns_mx_record" "testmx" {
    name                = "testmx"
    zone_name           = azurerm_private_dns_zone.testzone.name
    resource_group_name = data.azurerm_resource_group.example.name
    ttl                 = 300
    record {
        preference = 30
        exchange   = "bkpmx.thisisatestusingtf.com"
    }
}
