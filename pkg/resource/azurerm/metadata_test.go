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
		azurerm.AzureNetworkSecurityGroupResourceType:  {resource.FlagDeepMode},
		azurerm.AzureLoadBalancerResourceType:          {},
		azurerm.AzurePrivateDNSZoneResourceType:        {resource.FlagDeepMode},
		azurerm.AzurePrivateDNSARecordResourceType:     {resource.FlagDeepMode},
		azurerm.AzurePrivateDNSAAAARecordResourceType:  {resource.FlagDeepMode},
		azurerm.AzurePrivateDNSCNameRecordResourceType: {resource.FlagDeepMode},
		azurerm.AzurePrivateDNSPTRRecordResourceType:   {resource.FlagDeepMode},
		azurerm.AzurePrivateDNSMXRecordResourceType:    {resource.FlagDeepMode},
		azurerm.AzurePrivateDNSSRVRecordResourceType:   {resource.FlagDeepMode},
		azurerm.AzurePrivateDNSTXTRecordResourceType:   {resource.FlagDeepMode},
		azurerm.AzureImageResourceType:                 {},
		azurerm.AzureSSHPublicKeyResourceType:          {resource.FlagDeepMode},
		azurerm.AzureLoadBalancerRuleResourceType:      {resource.FlagDeepMode},
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
