package aws

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform/flatmap"
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
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

	results := make([]*resource.Resource, 0, len(stacks))

	for _, stack := range stacks {
		attrs := map[string]interface{}{}
		if len(stack.Parameters) > 0 {
			attrs["parameters.%"] = strconv.FormatInt(int64(len(stack.Parameters)), 10)
			for k, v := range flattenParameters(stack.Parameters) {
				attrs[fmt.Sprintf("parameters.%s", k)] = v
			}
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

func flattenParameters(parameters []*cloudformation.Parameter) flatmap.Map {
	params := make(map[string]interface{}, len(parameters))
	for _, p := range parameters {
		params[*p.ParameterKey] = *p.ParameterValue
	}
	return flatmap.Flatten(params)
}
