package aws

import (
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type AppAutoscalingTargetEnumerator struct {
	repository repository.AppAutoScalingRepository
	factory    resource.ResourceFactory
}

func NewAppAutoscalingTargetEnumerator(repository repository.AppAutoScalingRepository, factory resource.ResourceFactory) *AppAutoscalingTargetEnumerator {
	return &AppAutoscalingTargetEnumerator{
		repository,
		factory,
	}
}

func (e *AppAutoscalingTargetEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsAppAutoscalingTargetResourceType
}

func (e *AppAutoscalingTargetEnumerator) Enumerate() ([]*resource.Resource, error) {
	targets := make([]*applicationautoscaling.ScalableTarget, 0)

	for _, ns := range e.repository.ServiceNamespaceValues() {
		results, err := e.repository.DescribeScalableTargets(ns)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}
		targets = append(targets, results...)
	}

	results := make([]*resource.Resource, 0, len(targets))

	for _, target := range targets {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*target.ResourceId,
				map[string]interface{}{
					"service_namespace":  *target.ServiceNamespace,
					"scalable_dimension": *target.ScalableDimension,
				},
			),
		)
	}

	return results, nil
}
