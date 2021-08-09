package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsAutoScalingGroupResourceType = "aws_autoscaling_group"

func initAwsAutoScalingGroupMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsAutoScalingGroupResourceType, func(res *resource.AbstractResource) map[string]string {
		return map[string]string{
			"id":   res.TerraformId(),
			"name": res.TerraformId(),
			// "instances": res.Attrs.GetSlice("instances"),
			// "tags": res.Attrs.GetSlice("tags"),
			// "tag": res.Attrs.GetSlice("tag"),
			// "availability_zones": res.Attrs.GetSlice("availability_zones"),
			// "suspended_processes": res.Attrs.GetSlice("suspended_processes"),
			// "health_check_grace_period": res.Attrs.GetInt("health_check_grace_period"),
			// "health_check_type": *res.Attrs.GetString("health_check_type"),
			// "arn": *res.Attrs.GetString("arn"),
			// "launch_configuration": *res.Attrs.GetString("launch_configuration"),
			// "vpc_zone_identifier": *res.Attrs.GetString("vpc_zone_identifier"),
			// "min_size": fmt.Sprintf("%d",*res.Attrs.GetInt("min_size")),
			// "max_size": fmt.Sprintf("%d",*res.Attrs.GetInt("max_size")),
			// "max_instance_lifetime": fmt.Sprintf("%d",*res.Attrs.GetInt("max_instance_lifetime")),
			// "placement_group": res.Attrs.Get("placement_group"),
			// "desired_capacity": res.Attrs.Get("desired_capacity"),
			// "service_linked_role_arn": res.Attrs.Get("service_linked_role_arn"),
			// "default_cooldown": res.Attrs.Get("default_cooldown"),
		}
	})
}
