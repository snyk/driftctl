package azurerm

import "github.com/cloudskiff/driftctl/pkg/resource"

func InitResourcesMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	initAzureVirtualNetworkMetaData(resourceSchemaRepository)
	initAzureRouteTableMetaData(resourceSchemaRepository)
}
