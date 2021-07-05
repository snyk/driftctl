package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	tf "github.com/cloudskiff/driftctl/pkg/remote/terraform"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type EC2InstanceEnumerator struct {
	repository     repository.EC2Repository
	factory        resource.ResourceFactory
	providerConfig tf.TerraformProviderConfig
}

func NewEC2InstanceEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory, providerConfig tf.TerraformProviderConfig) *EC2InstanceEnumerator {
	return &EC2InstanceEnumerator{
		repository:     repo,
		factory:        factory,
		providerConfig: providerConfig,
	}
}

func (e *EC2InstanceEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsInstanceResourceType
}

func (e *EC2InstanceEnumerator) Enumerate() ([]resource.Resource, error) {
	instances, err := e.repository.ListAllInstances()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, string(e.SupportedType()))
	}

	results := make([]resource.Resource, len(instances))

	for _, instance := range instances {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*instance.InstanceId,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
