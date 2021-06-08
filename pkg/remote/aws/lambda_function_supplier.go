package aws

import (
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type LambdaFunctionSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	client       repository.LambdaRepository
	runner       *terraform.ParallelResourceReader
}

func NewLambdaFunctionSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.LambdaRepository) *LambdaFunctionSupplier {
	return &LambdaFunctionSupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *LambdaFunctionSupplier) SuppliedType() resource.ResourceType {
	return resourceaws.AwsLambdaFunctionResourceType
}

func (s *LambdaFunctionSupplier) Resources() ([]resource.Resource, error) {
	functions, err := s.client.ListAllLambdaFunctions()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
	}
	results := make([]cty.Value, 0)
	if len(functions) > 0 {
		for _, function := range functions {
			fun := *function
			s.runner.Run(func() (cty.Value, error) {
				return s.readLambda(fun)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(s.SuppliedType(), results)
}

func (s *LambdaFunctionSupplier) readLambda(function lambda.FunctionConfiguration) (cty.Value, error) {
	name := *function.FunctionName
	resFunction, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: s.SuppliedType(),
			ID: name,
			Attributes: map[string]string{
				"function_name": name,
			},
		},
	)
	if err != nil {
		logrus.Warnf("Error reading function %s[%s]: %+v", name, s.SuppliedType(), err)
		return cty.NilVal, err
	}

	return *resFunction, nil
}
