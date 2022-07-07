package google

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/google"
)

func initGoogleComputeFirewallMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(google.GoogleComputeFirewallResourceType, func(res *resource.Resource) {
		res.Attrs.SafeDelete([]string{"timeouts"})
	})
}
