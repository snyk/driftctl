package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type EC2EipSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	client       repository.EC2Repository
	runner       *terraform.ParallelResourceReader
}

func NewEC2EipSupplier(provider *AWSTerraformProvider, repo repository.EC2Repository, deserializer *resource.Deserializer) *EC2EipSupplier {
	return &EC2EipSupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *EC2EipSupplier) SuppliedType() resource.ResourceType {
	return resourceaws.AwsEipResourceType
}

func (s *EC2EipSupplier) Resources() ([]resource.Resource, error) {
	addresses, err := s.client.ListAllAddresses()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
	}
	results := make([]cty.Value, 0)
	if len(addresses) > 0 {
		for _, address := range addresses {
			addr := *address
			s.runner.Run(func() (cty.Value, error) {
				return s.readEIP(addr)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(s.SuppliedType(), results)
}

func (s *EC2EipSupplier) readEIP(address ec2.Address) (cty.Value, error) {
	id := aws.StringValue(address.AllocationId)
	resAddress, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: s.SuppliedType(),
		ID: id,
	})
	if err != nil {
		logrus.Warnf("Error reading eip %s[%s]: %+v", id, s.SuppliedType(), err)
		return cty.NilVal, err
	}
	return *resAddress, nil
}
