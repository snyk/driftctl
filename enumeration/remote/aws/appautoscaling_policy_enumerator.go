package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
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
	results := make([]*resource.Resource, 0)

	for _, ns := range e.repository.ServiceNamespaceValues() {
		policies, err := e.repository.DescribeScalingPolicies(ns)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}

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
	}

	return results, nil
}
