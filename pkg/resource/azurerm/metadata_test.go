package azurerm_test

import (
	"testing"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/azurerm"
	testresource "github.com/snyk/driftctl/test/resource"
	"github.com/stretchr/testify/assert"
)

func TestAzureMetadata_Flags(t *testing.T) {
	testcases := map[string][]resource.Flags{
		azurerm.AzureContainerRegistryResourceType:     {},
		azurerm.AzureFirewallResourceType:              {},
		azurerm.AzurePostgresqlServerResourceType:      {},
		azurerm.AzurePostgresqlDatabaseResourceType:    {},
		azurerm.AzurePublicIPResourceType:              {},
		azurerm.AzureResourceGroupResourceType:         {},
		azurerm.AzureRouteResourceType:                 {},
		azurerm.AzureRouteTableResourceType:            {},
		azurerm.AzureStorageAccountResourceType:        {},
		azurerm.AzureStorageContainerResourceType:      {},
		azurerm.AzureSubnetResourceType:                {},
		azurerm.AzureVirtualNetworkResourceType:        {},
		azurerm.AzureNetworkSecurityGroupResourceType:  {},
		azurerm.AzureLoadBalancerResourceType:          {},
		azurerm.AzurePrivateDNSZoneResourceType:        {},
		azurerm.AzurePrivateDNSARecordResourceType:     {},
		azurerm.AzurePrivateDNSAAAARecordResourceType:  {},
		azurerm.AzurePrivateDNSCNameRecordResourceType: {},
		azurerm.AzurePrivateDNSPTRRecordResourceType:   {},
		azurerm.AzurePrivateDNSMXRecordResourceType:    {},
		azurerm.AzurePrivateDNSSRVRecordResourceType:   {},
		azurerm.AzurePrivateDNSTXTRecordResourceType:   {},
		azurerm.AzureImageResourceType:                 {},
		azurerm.AzureSSHPublicKeyResourceType:          {},
		azurerm.AzureLoadBalancerRuleResourceType:      {},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("azurerm", "2.71.0")
	azurerm.InitResourcesMetadata(schemaRepository)

	for ty, flags := range testcases {
		t.Run(ty, func(tt *testing.T) {
			sch, exist := schemaRepository.GetSchema(ty)
			assert.True(tt, exist)

			if len(flags) == 0 {
				assert.Equal(tt, resource.Flags(0x0), sch.Flags, "should not have any flag")
				return
			}

			for _, flag := range flags {
				assert.Truef(tt, sch.Flags.HasFlag(flag), "should have given flag %d", flag)
			}
		})
	}
}
