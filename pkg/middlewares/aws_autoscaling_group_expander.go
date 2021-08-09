package middlewares

import (
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type AwsAutoScalingGroupInstanceExpander struct {
}

func NewAutoScalingGroupInstanceExpander() *AwsAutoScalingGroupInstanceExpander {
	return &AwsAutoScalingGroupInstanceExpander{}
}

func (a AwsAutoScalingGroupInstanceExpander) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {

	remoteInstances := make([]resource.Resource, 0)

	for _, res := range *remoteResources {
		// Ignore all resources other than aws_instance
		if res.TerraformType() != aws.AwsInstanceResourceType {
			continue
		}
		remoteInstances = append(remoteInstances, res)
	}

	for _, res := range *remoteResources {
		// Ignore all resources other than aws_autoscaling_group
		if res.TerraformType() != aws.AwsAutoScalingGroupResourceType {
			continue
		}

		instances, exist := res.Attributes().Get("instances")
		if !exist {
			continue
		}

		for _, item := range instances.([]*autoscaling.Instance) {
			for _, instance := range remoteInstances {
				if instance.TerraformId() == *item.InstanceId {
					*resourcesFromState = append(*resourcesFromState, instance)
					break
				}
			}
		}
	}

	return nil
}
