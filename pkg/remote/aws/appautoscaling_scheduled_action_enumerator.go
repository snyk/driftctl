package aws

import (
	"strings"

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
	results := make([]*resource.Resource, 0)

	for _, ns := range e.repository.ServiceNamespaceValues() {
		actions, err := e.repository.DescribeScheduledActions(ns)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}

		for _, action := range actions {
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					strings.Join([]string{*action.ScheduledActionName, *action.ServiceNamespace, *action.ResourceId}, "-"),
					map[string]interface{}{},
				),
			)
		}
	}

	return results, nil
}
