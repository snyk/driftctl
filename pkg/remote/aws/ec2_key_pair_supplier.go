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

type EC2KeyPairSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	client       repository.EC2Repository
	runner       *terraform.ParallelResourceReader
}

func NewEC2KeyPairSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer) *EC2KeyPairSupplier {
	return &EC2KeyPairSupplier{
		provider,
		deserializer,
		repository.NewEC2Repository(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *EC2KeyPairSupplier) Resources() ([]resource.Resource, error) {
	keyPairs, err := s.client.ListAllKeyPairs()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsKeyPairResourceType)
	}
	results := make([]cty.Value, 0)
	if len(keyPairs) > 0 {
		for _, kp := range keyPairs {
			name := aws.StringValue(kp.KeyName)
			s.runner.Run(func() (cty.Value, error) {
				return s.readKeyPair(name)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(resourceaws.AwsKeyPairResourceType, results)
}

func (s *EC2KeyPairSupplier) readKeyPair(name string) (cty.Value, error) {
	resKp, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: resourceaws.AwsKeyPairResourceType,
		ID: name,
	})
	if err != nil {
		logrus.Warnf("Error reading key pair %s: %+v", name, err)
		return cty.NilVal, err
	}
	return *resKp, nil
}
