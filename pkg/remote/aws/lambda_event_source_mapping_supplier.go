package aws

import (
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type LambdaEventSourceMappingSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       repository.LambdaRepository
	runner       *terraform.ParallelResourceReader
}

func NewLambdaEventSourceMappingSupplier(provider *AWSTerraformProvider) *LambdaEventSourceMappingSupplier {
	return &LambdaEventSourceMappingSupplier{
		provider,
		awsdeserializer.NewLambdaEventSourceMappingDeserializer(),
		repository.NewLambdaRepository(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *LambdaEventSourceMappingSupplier) Resources() ([]resource.Resource, error) {
	functions, err := s.client.ListAllLambdaEventSourceMappings()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsLambdaEventSourceMappingResourceType)
	}
	var results []cty.Value
	for _, function := range functions {
		fun := *function
		s.runner.Run(func() (cty.Value, error) {
			return s.readLambdaEventSourceMapping(fun)
		})
	}
	results, err = s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(results)
}

func (s *LambdaEventSourceMappingSupplier) readLambdaEventSourceMapping(sourceMappingConfig lambda.EventSourceMappingConfiguration) (cty.Value, error) {
	resFunction, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: resourceaws.AwsLambdaEventSourceMappingResourceType,
			ID: *sourceMappingConfig.UUID,
		},
	)
	if err != nil {
		logrus.Warnf("Error reading %s: %+v", resourceaws.AwsLambdaEventSourceMappingResourceType, err)
		return cty.NilVal, err
	}

	return *resFunction, nil
}
