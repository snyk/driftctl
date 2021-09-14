package aws

import (
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
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

	results := make([]*resource.Resource, len(targets))

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
