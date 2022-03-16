package google

import (
	"testing"

	"github.com/snyk/driftctl/pkg/resource"
	tf "github.com/snyk/driftctl/pkg/terraform"
	testresource "github.com/snyk/driftctl/test/resource"
	"github.com/stretchr/testify/assert"
)

func TestGoogle_Metadata_Flags(t *testing.T) {
	testcases := map[string][]resource.Flags{
		GoogleBigqueryDatasetResourceType:         {},
		GoogleComputeFirewallResourceType:         {resource.FlagDeepMode},
		GoogleComputeInstanceResourceType:         {},
		GoogleComputeInstanceGroupResourceType:    {resource.FlagDeepMode},
		GoogleComputeNetworkResourceType:          {resource.FlagDeepMode},
		GoogleComputeRouterResourceType:           {},
		GoogleDNSManagedZoneResourceType:          {},
		GoogleProjectIamBindingResourceType:       {},
		GoogleProjectIamMemberResourceType:        {resource.FlagDeepMode},
		GoogleProjectIamPolicyResourceType:        {},
		GoogleStorageBucketResourceType:           {resource.FlagDeepMode},
		GoogleStorageBucketIamBindingResourceType: {},
		GoogleStorageBucketIamMemberResourceType:  {resource.FlagDeepMode},
		GoogleStorageBucketIamPolicyResourceType:  {},
		GoogleBigqueryTableResourceType:           {},
		GoogleComputeDiskResourceType:             {},
		GoogleBigTableInstanceResourceType:        {},
		GoogleComputeGlobalAddressResourceType:    {},
		GoogleCloudRunServiceResourceType:         {},
		GoogleComputeNodeGroupResourceType:        {},
		GoogleComputeForwardingRuleResourceType:   {},
	}

	schemaRepository := testresource.InitFakeSchemaRepository(tf.GOOGLE, "3.78.0")
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
