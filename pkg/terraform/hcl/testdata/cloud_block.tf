terraform {
    cloud {
        organization = "example_corp"
        hostname = "app.terraform.io" # Optional; defaults to app.terraform.io

        workspaces {}
    }
}
