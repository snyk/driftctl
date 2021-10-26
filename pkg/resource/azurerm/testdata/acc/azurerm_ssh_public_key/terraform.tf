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

resource "azurerm_ssh_public_key" "example" {
    name                = "acc-test-key"
    resource_group_name = data.azurerm_resource_group.qa1.name
    location            = data.azurerm_resource_group.qa1.location
    public_key          = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDzBNA813NC+4myQMPWZpXFzbyWkHzZMET7Tu+ZOo5b9GkTmh/d5LvXZrKGy4YCh/Wuknwfrlg6b2EDJdm5DOV8H61dX2g/UfMYKLyczD+cdIOyDCxAU4Hj+JyIg+KaZJN1kikVlm6XhnZfMipE7z1F28VKYoro9+3Nt/mg4+/lCWp/0a6Bkh7q1V4EXO3x2yA39jqbmMUylnzD0EuBnECmTBy9aCUR7vAMcKSPgG9Z6RD2+COVtdz/fmWKI8P02Pocv7Sl5EcvbN+sTfnFavFMcbQMcgM4oPSB1CNg/jWn6dZh2Wb04n1kpnWHe+q/1UEwKtKHcT3hQH2I+Ip45EgIEpXpRcUuOYf+8wHfml1CM9gy84QYQ0Rqy9Rhr6BAYg6XzE/FjxOoarRxoN/D8Z0Ld3hXqk09pzUbjC/b2hSzgALsVUvYfM2Q0/Vj7ufKMRxqv5vlCNmM4/LJGlxethl+zFkwl/JucKhjLDNNNoUANVp3QPNCztyrFBfBUYYCii5p3SBuCUUJ63a0m/nCt8frRZjzTmbCel1jiQDehOCJQ1lmIQthAKUtYNYkN5vRjhWa6CoobHeWOYS48QCTMABkFq7ewTW6H/LyWaRa5/34Z1b9K9Ht53oCkQOzSYaDp+XZZ+lvTD0/4ArmFGqzeKVi7AExJUlbSQd5stjLixA6mQ== acc@driftctl.com"
}
