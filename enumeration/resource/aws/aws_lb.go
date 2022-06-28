package aws

import "github.com/snyk/driftctl/enumeration/resource"

const AwsLoadBalancerResourceType = "aws_lb"

func initAwsLoadBalancerMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsLoadBalancerResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"Name": *res.Attributes().GetString("name"),
		}
	})
}
