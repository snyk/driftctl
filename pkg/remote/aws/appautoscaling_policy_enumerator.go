package aws

import (
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type AppAutoscalingPolicyEnumerator struct {
	repository repository.AppAutoScalingRepository
	factory    resource.ResourceFactory
}

func NewAppAutoscalingPolicyEnumerator(repository repository.AppAutoScalingRepository, factory resource.ResourceFactory) *AppAutoscalingPolicyEnumerator {
	return &AppAutoscalingPolicyEnumerator{
		repository,
		factory,
	}
}

func (e *AppAutoscalingPolicyEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsAppAutoscalingPolicyResourceType
}

func (e *AppAutoscalingPolicyEnumerator) Enumerate() ([]*resource.Resource, error) {
	policies := make([]*applicationautoscaling.ScalingPolicy, 0)

	for _, ns := range e.repository.ServiceNamespaceValues() {
		results, err := e.repository.DescribeScalingPolicies(ns)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}
		policies = append(policies, results...)
	}

	results := make([]*resource.Resource, 0)

	for _, policy := range policies {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*policy.PolicyName,
				map[string]interface{}{
					"name":               *policy.PolicyName,
					"resource_id":        *policy.ResourceId,
					"scalable_dimension": *policy.ScalableDimension,
					"service_namespace":  *policy.ServiceNamespace,
				},
			),
		)
	}

	return results, nil
}
