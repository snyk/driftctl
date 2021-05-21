package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsRoute53ZoneResourceType = "aws_route53_zone"

func initAwsRoute53ZoneMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsRoute53ZoneResourceType, func(res *resource.AbstractResource) {
		val := res.Attrs
		val.SafeDelete([]string{"force_destroy"})
	})
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsRoute53ZoneResourceType, func(res *resource.AbstractResource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if name := val.GetString("name"); name != nil && *name != "" {
			attrs["Name"] = *name
		}
		return attrs
	})
}
