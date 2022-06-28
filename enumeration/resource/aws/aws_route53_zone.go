package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsRoute53ZoneResourceType = "aws_route53_zone"

func initAwsRoute53ZoneMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsRoute53ZoneResourceType, func(res *resource.Resource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if name := val.GetString("name"); name != nil && *name != "" {
			attrs["Name"] = *name
		}
		return attrs
	})
	resourceSchemaRepository.SetFlags(AwsRoute53ZoneResourceType, resource.FlagDeepMode)
}
