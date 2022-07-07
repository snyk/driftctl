package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"strings"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
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
