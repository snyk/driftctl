package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type EC2AmiSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       repository.EC2Repository
	runner       *terraform.ParallelResourceReader
}

func NewEC2AmiSupplier(provider *AWSTerraformProvider) *EC2AmiSupplier {
	return &EC2AmiSupplier{
		provider,
		awsdeserializer.NewEC2AmiDeserializer(),
		repository.NewEC2Repository(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s EC2AmiSupplier) Resources() ([]resource.Resource, error) {
	images, err := s.client.ListAllImages()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsAmiResourceType)
	}
	results := make([]cty.Value, 0)
	if len(images) > 0 {
		for _, image := range images {
			id := aws.StringValue(image.ImageId)
			s.runner.Run(func() (cty.Value, error) {
				return s.readAMI(id)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(results)
}

func (s EC2AmiSupplier) readAMI(id string) (cty.Value, error) {
	resImage, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: resourceaws.AwsAmiResourceType,
		ID: id,
	})
	if err != nil {
		logrus.Warnf("Error reading image %s[%s]: %+v", id, resourceaws.AwsAmiResourceType, err)
		return cty.NilVal, err
	}
	return *resImage, nil
}
