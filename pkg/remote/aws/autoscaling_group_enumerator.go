package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

type AutoScalingGroupsEnumerator struct {
	repo    repository.AutoScalingRepository
	factory resource.ResourceFactory
}

func NewAutoScalingGroupsEnumerator(repo repository.AutoScalingRepository, factory resource.ResourceFactory) *AutoScalingGroupsEnumerator {
	return &AutoScalingGroupsEnumerator{
		repo,
		factory,
	}
}

func (e *AutoScalingGroupsEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsAutoScalingGroupResourceType
}

func (e *AutoScalingGroupsEnumerator) Enumerate() ([]resource.Resource, error) {
	groups, err := e.repo.DescribeGroups([]*string{})
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]resource.Resource, 0, len(groups))

	for _, item := range groups {
		var attrs = map[string]interface{}{
			"instances":           item.Instances,
			"tags":                item.Tags,
			"tag":                 item.Tags,
			"availability_zones":  item.AvailabilityZones,
			"suspended_processes": item.SuspendedProcesses,
		}

		if val := item.HealthCheckGracePeriod; val != nil {
			attrs["health_check_grace_period"] = *val
		}
		if val := item.HealthCheckType; val != nil {
			attrs["health_check_type"] = *val
		}
		if val := item.AutoScalingGroupARN; val != nil {
			attrs["arn"] = *val
		}
		if val := item.LaunchConfigurationName; val != nil {
			attrs["launch_configuration"] = *val
		}
		if val := item.VPCZoneIdentifier; val != nil {
			attrs["vpc_zone_identifier"] = *val
		}
		if val := item.MinSize; val != nil {
			attrs["min_size"] = *val
		}
		if val := item.MaxSize; val != nil {
			attrs["max_size"] = *val
		}
		if val := item.MaxInstanceLifetime; val != nil {
			attrs["max_instance_lifetime"] = *val
		}
		if val := item.PlacementGroup; val != nil {
			attrs["placement_group"] = *val
		}
		if val := item.DesiredCapacity; val != nil {
			attrs["desired_capacity"] = *val
		}
		if val := item.ServiceLinkedRoleARN; val != nil {
			attrs["service_linked_role_arn"] = *val
		}
		if val := item.DefaultCooldown; val != nil {
			attrs["default_cooldown"] = *val
		}

		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*item.AutoScalingGroupName,
				attrs,
			),
		)
	}

	return results, nil
}
