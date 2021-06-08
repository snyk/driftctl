package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type EC2InstanceSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	client       repository.EC2Repository
	runner       *terraform.ParallelResourceReader
}

func NewEC2InstanceSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.EC2Repository) *EC2InstanceSupplier {
	return &EC2InstanceSupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *EC2InstanceSupplier) SuppliedType() resource.ResourceType {
	return resourceaws.AwsInstanceResourceType
}

func (s *EC2InstanceSupplier) Resources() ([]resource.Resource, error) {
	instances, err := s.client.ListAllInstances()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
	}

	results := make([]cty.Value, 0)
	if len(instances) > 0 {
		for _, instance := range instances {
			id := aws.StringValue(instance.InstanceId)
			s.runner.Run(func() (cty.Value, error) {
				return s.readInstance(id)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(s.SuppliedType(), results)
}

func (s *EC2InstanceSupplier) readInstance(id string) (cty.Value, error) {
	resInstance, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: s.SuppliedType(),
		ID: id,
	})
	if err != nil {
		logrus.Warnf("Error reading instance %s: %+v", id, err)
		return cty.NilVal, err
	}
	if resInstance.IsNull() {
		logrus.WithFields(logrus.Fields{
			"id": id,
		}).Debug("Instance read returned nil (instance may be terminated), ignoring ...")
	}
	return *resInstance, nil
}
