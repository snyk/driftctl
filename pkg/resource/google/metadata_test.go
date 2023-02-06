package google_test

import (
	"testing"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/google"
	testresource "github.com/snyk/driftctl/test/resource"
	"github.com/stretchr/testify/assert"
)

func TestGoogle_Metadata_Flags(t *testing.T) {
	testcases := map[string][]resource.Flags{
		google.GoogleBigqueryDatasetResourceType:             {},
		google.GoogleComputeFirewallResourceType:             {resource.FlagDeepMode},
		google.GoogleComputeInstanceResourceType:             {},
		google.GoogleComputeInstanceGroupResourceType:        {resource.FlagDeepMode},
		google.GoogleComputeNetworkResourceType:              {resource.FlagDeepMode},
		google.GoogleComputeRouterResourceType:               {},
		google.GoogleDNSManagedZoneResourceType:              {},
		google.GoogleProjectIamBindingResourceType:           {},
		google.GoogleProjectIamMemberResourceType:            {resource.FlagDeepMode},
		google.GoogleProjectIamPolicyResourceType:            {},
		google.GoogleStorageBucketResourceType:               {resource.FlagDeepMode},
		google.GoogleStorageBucketIamBindingResourceType:     {},
		google.GoogleStorageBucketIamMemberResourceType:      {resource.FlagDeepMode},
		google.GoogleStorageBucketIamPolicyResourceType:      {},
		google.GoogleBigqueryTableResourceType:               {},
		google.GoogleComputeDiskResourceType:                 {},
		google.GoogleBigTableInstanceResourceType:            {},
		google.GoogleComputeGlobalAddressResourceType:        {},
		google.GoogleCloudRunServiceResourceType:             {},
		google.GoogleComputeNodeGroupResourceType:            {},
		google.GoogleComputeForwardingRuleResourceType:       {},
		google.GoogleComputeInstanceGroupManagerResourceType: {},
		google.GoogleComputeGlobalForwardingRuleResourceType: {},
		google.GoogleComputeSslCertificateResourceType:       {},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("google", "3.78.0")
	google.InitResourcesMetadata(schemaRepository)

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
