package aws

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type CloudformationStackEnumerator struct {
	repository repository.CloudformationRepository
	factory    resource.ResourceFactory
}

func NewCloudformationStackEnumerator(repo repository.CloudformationRepository, factory resource.ResourceFactory) *CloudformationStackEnumerator {
	return &CloudformationStackEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *CloudformationStackEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsCloudformationStackResourceType
}

func (e *CloudformationStackEnumerator) Enumerate() ([]*resource.Resource, error) {
	stacks, err := e.repository.ListAllStacks()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, len(stacks))

	for _, stack := range stacks {
		attrs := map[string]interface{}{}
		if stack.Parameters != nil && len(stack.Parameters) > 0 {
			attrs["parameters"] = flattenParameters(stack.Parameters)
		}

		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*stack.StackId,
				attrs,
			),
		)
	}

	return results, err
}

func flattenParameters(parameters []*cloudformation.Parameter) interface{} {
	params := make(map[string]interface{}, len(parameters))
	for _, p := range parameters {
		params[*p.ParameterKey] = *p.ParameterValue
	}
	return params
}
