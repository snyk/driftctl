package aws

import (
	"fmt"

	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsRoute53HealthCheckResourceType = "aws_route53_health_check"

func initAwsRoute53HealthCheckMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsRoute53HealthCheckResourceType, func(res *resource.Resource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if tags := val.GetMap("tags"); tags != nil {
			if name, ok := tags["Name"]; ok {
				attrs["Name"] = name.(string)
			}
		}
		port := val.GetInt("port")
		path := val.GetString("resource_path")
		if fqdn := val.GetString("fqdn"); fqdn != nil && *fqdn != "" {
			attrs["Fqdn"] = *fqdn
			if port != nil {
				attrs["Port"] = fmt.Sprintf("%d", *port)
			}
			if path != nil && *path != "" {
				attrs["Path"] = *path
			}
		}
		if address := val.GetString("ip_address"); address != nil && *address != "" {
			attrs["IpAddress"] = *address
			if port != nil {
				attrs["Port"] = fmt.Sprintf("%d", *port)
			}
			if path != nil && *path != "" {
				attrs["Path"] = *path
			}
		}
		return attrs
	})
}
