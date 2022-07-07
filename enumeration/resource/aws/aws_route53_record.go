package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsRoute53RecordResourceType = "aws_route53_record"

func initAwsRoute53RecordMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsRoute53RecordResourceType, func(res *resource.Resource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if fqdn := val.GetString("fqdn"); fqdn != nil && *fqdn != "" {
			attrs["Fqdn"] = *fqdn
		}
		if ty := val.GetString("type"); ty != nil && *ty != "" {
			attrs["Type"] = *ty
		}
		if zoneID := val.GetString("zone_id"); zoneID != nil && *zoneID != "" {
			attrs["ZoneId"] = *zoneID
		}
		return attrs
	})
	resourceSchemaRepository.SetFlags(AwsRoute53RecordResourceType, resource.FlagDeepMode)
}
