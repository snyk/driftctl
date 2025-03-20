provider "scaleway" {}

terraform {
    required_providers {
        scaleway = {
            source  = "scaleway/scaleway"
            version = "2.14.1"
        }
    }
}

resource "scaleway_function_namespace" "namespace" {
    name        = "TestAcc-Scaleway-FunctionNamespace"
    description = "This is a test description"
}
