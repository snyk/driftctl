package terraform

import tfproviders "github.com/hashicorp/terraform/providers"

type SchemaSupplier interface {
	Schema() map[string]tfproviders.Schema
}
