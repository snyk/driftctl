package azurerm

import (
	"testing"

	"github.com/snyk/driftctl/pkg/resource"
	tf "github.com/snyk/driftctl/pkg/terraform"
	testresource "github.com/snyk/driftctl/test/resource"
	"github.com/stretchr/testify/assert"
)

func TestAzureMetadata_Flags(t *testing.T) {
	testcases := map[string][]resource.Flags{
		AzureContainerRegistryResourceType:     {},
		AzureFirewallResourceType:              {},
		AzurePostgresqlServerResourceType:      {},
		AzurePostgresqlDatabaseResourceType:    {},
		AzurePublicIPResourceType:              {},
		AzureResourceGroupResourceType:         {},
		AzureRouteResourceType:                 {},
		AzureRouteTableResourceType:            {},
		AzureStorageAccountResourceType:        {},
		AzureStorageContainerResourceType:      {},
		AzureSubnetResourceType:                {},
		AzureVirtualNetworkResourceType:        {},
		AzureNetworkSecurityGroupResourceType:  {resource.FlagDeepMode},
		AzureLoadBalancerResourceType:          {},
		AzurePrivateDNSZoneResourceType:        {resource.FlagDeepMode},
		AzurePrivateDNSARecordResourceType:     {resource.FlagDeepMode},
		AzurePrivateDNSAAAARecordResourceType:  {resource.FlagDeepMode},
		AzurePrivateDNSCNameRecordResourceType: {resource.FlagDeepMode},
		AzurePrivateDNSPTRRecordResourceType:   {resource.FlagDeepMode},
		AzurePrivateDNSMXRecordResourceType:    {resource.FlagDeepMode},
		AzurePrivateDNSSRVRecordResourceType:   {resource.FlagDeepMode},
		AzurePrivateDNSTXTRecordResourceType:   {resource.FlagDeepMode},
		AzureImageResourceType:                 {},
		AzureSSHPublicKeyResourceType:          {resource.FlagDeepMode},
		AzureLoadBalancerRuleResourceType:      {resource.FlagDeepMode},
	}

	schemaRepository := testresource.InitFakeSchemaRepository(tf.AZURE, "2.71.0")
	InitResourcesMetadata(schemaRepository)

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
