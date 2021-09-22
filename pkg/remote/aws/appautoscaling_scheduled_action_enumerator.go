package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type AppAutoscalingScheduledActionEnumerator struct {
	repository repository.AppAutoScalingRepository
	factory    resource.ResourceFactory
}

func NewAppAutoscalingScheduledActionEnumerator(repository repository.AppAutoScalingRepository, factory resource.ResourceFactory) *AppAutoscalingScheduledActionEnumerator {
	return &AppAutoscalingScheduledActionEnumerator{
		repository,
		factory,
	}
}

func (e *AppAutoscalingScheduledActionEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsAppAutoscalingScheduledActionResourceType
}

func (e *AppAutoscalingScheduledActionEnumerator) Enumerate() ([]*resource.Resource, error) {
	actions := make([]*applicationautoscaling.ScheduledAction, 0)

	for _, ns := range e.repository.ServiceNamespaceValues() {
		results, err := e.repository.DescribeScheduledActions(ns)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}
		actions = append(actions, results...)
	}

	results := make([]*resource.Resource, len(actions))

	for _, action := range actions {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				fmt.Sprintf("%s-%s-%s", *action.ScheduledActionName, *action.ServiceNamespace, *action.ResourceId),
				map[string]interface{}{
					"name":               *action.ScheduledActionName,
					"service_namespace":  *action.ServiceNamespace,
					"scalable_dimension": *action.ScalableDimension,
					"resource_id":        *action.ResourceId,
				},
			),
		)
	}

	return results, nil
}
