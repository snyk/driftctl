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

resource "azurerm_private_dns_aaaa_record" "othertestaaaa" {
    name                = "othertest"
    zone_name           = azurerm_private_dns_zone.testzone.name
    resource_group_name = data.azurerm_resource_group.example.name
    ttl                 = 300
    records             = ["fd5d:70bc:930e:d008:0000:0000:0000:7334", "fd5d:70bc:930e:d008::7335"]
}
resource "azurerm_private_dns_aaaa_record" "testaaaa" {
    name                = "test"
    zone_name           = azurerm_private_dns_zone.testzone.name
    resource_group_name = data.azurerm_resource_group.example.name
    ttl                 = 300
    records             = ["fd5d:70bc:930e:d008:0000:0000:0000:7334", "fd5d:70bc:930e:d008::7335"]
}
