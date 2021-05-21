package aws

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	testresource "github.com/cloudskiff/driftctl/test/resource"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/stretchr/testify/assert"

	"github.com/cloudskiff/driftctl/pkg/parallel"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/mocks"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
)

func TestLambdaFunctionSupplier_Resources(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repo *repository.MockLambdaRepository)
		err     error
	}{
		{
			test:    "no lambda functions",
			dirName: "lambda_function_empty",
			mocks: func(repo *repository.MockLambdaRepository) {
				repo.On("ListAllLambdaFunctions").Return([]*lambda.FunctionConfiguration{}, nil)
			},
			err: nil,
		},
		{
			test:    "with lambda functions",
			dirName: "lambda_function_multiple",
			mocks: func(repo *repository.MockLambdaRepository) {
				repo.On("ListAllLambdaFunctions").Return([]*lambda.FunctionConfiguration{
					{
						FunctionName: aws.String("foo"),
					},
					{
						FunctionName: aws.String("bar"),
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "One lambda with signing",
			dirName: "lambda_function_signed",
			mocks: func(repo *repository.MockLambdaRepository) {
				repo.On("ListAllLambdaFunctions").Return([]*lambda.FunctionConfiguration{
					{
						FunctionName: aws.String("foo"),
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list lambda functions",
			dirName: "lambda_function_empty",
			mocks: func(repo *repository.MockLambdaRepository) {
				repo.On("ListAllLambdaFunctions").Return([]*lambda.FunctionConfiguration{}, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsLambdaFunctionResourceType),
		},
	}
	for _, tt := range tests {
		shouldUpdate := tt.dirName == *goldenfile.Update

		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		repo := testresource.InitFakeSchemaRepository("aws", "3.19.0")
		resourceaws.InitResourcesMetadata(repo)
		factory := terraform.NewTerraformResourceFactory(repo)

		deserializer := resource.NewDeserializer(factory)
		if shouldUpdate {
			provider, err := InitTestAwsProvider(providerLibrary)
			if err != nil {
				t.Fatal(err)
			}
			supplierLibrary.AddSupplier(NewLambdaFunctionSupplier(provider, deserializer))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			client := &repository.MockLambdaRepository{}
			tt.mocks(client)
			s := &LambdaFunctionSupplier{
				provider,
				deserializer,
				client,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(t, tt.err, err)

			test.CtyTestDiff(got, tt.dirName, provider, deserializer, shouldUpdate, t)
		})
	}
}
