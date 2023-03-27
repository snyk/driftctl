package scaleway_test

import (
	"testing"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/scaleway"
	testresource "github.com/snyk/driftctl/test/resource"
	"github.com/stretchr/testify/assert"
)

func TestScaleway_Metadata_Flags(t *testing.T) {
	testcases := map[string][]resource.Flags{
		scaleway.ScalewayFunctionNamespaceResourceType: {resource.FlagDeepMode},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("scaleway", "2.14.1")
	scaleway.InitResourcesMetadata(schemaRepository)

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
