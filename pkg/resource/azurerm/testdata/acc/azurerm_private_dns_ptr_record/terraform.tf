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

resource "azurerm_private_dns_ptr_record" "othertestptr" {
    name                = "othertestptr"
    zone_name           = azurerm_private_dns_zone.testzone.name
    resource_group_name = data.azurerm_resource_group.example.name
    ttl                 = 300
    records = ["ptr1.thisisatestusingtf.com", "ptr2.thisisatestusingtf.com"]
}

resource "azurerm_private_dns_ptr_record" "testptr" {
    name                = "testptr"
    zone_name           = azurerm_private_dns_zone.testzone.name
    resource_group_name = data.azurerm_resource_group.example.name
    ttl                 = 300
    records = ["ptr3.thisisatestusingtf.com"]

}
