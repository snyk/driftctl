package aws

import (
	"context"
	"testing"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/stretchr/testify/assert"

	"github.com/cloudskiff/driftctl/pkg/parallel"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

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
		test           string
		dirName        string
		functionsPages mocks.ListFunctionsPagesOutput
		listError      error
		err            error
	}{
		{
			test:    "no lambda functions",
			dirName: "lambda_function_empty",
			functionsPages: mocks.ListFunctionsPagesOutput{
				{
					true,
					&lambda.ListFunctionsOutput{},
				},
			},
			err: nil,
		},
		{
			test:    "with lambda functions",
			dirName: "lambda_function_multiple",
			functionsPages: mocks.ListFunctionsPagesOutput{
				{
					false,
					&lambda.ListFunctionsOutput{
						Functions: []*lambda.FunctionConfiguration{
							{
								FunctionName: aws.String("foo"),
							},
						},
					},
				},
				{
					true,
					&lambda.ListFunctionsOutput{
						Functions: []*lambda.FunctionConfiguration{
							{
								FunctionName: aws.String("bar"),
							},
						},
					},
				},
			},
			err: nil,
		},
		{
			test:    "One lambda with signing",
			dirName: "lambda_function_signed",
			functionsPages: mocks.ListFunctionsPagesOutput{
				{
					false,
					&lambda.ListFunctionsOutput{
						Functions: []*lambda.FunctionConfiguration{
							{
								FunctionName: aws.String("foo"),
							},
						},
					},
				},
			},
			err: nil,
		},
		{
			test:      "cannot list lambda functions",
			dirName:   "lambda_function_empty",
			listError: awserr.NewRequestFailure(nil, 403, ""),
			err:       remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsLambdaFunctionResourceType),
		},
	}
	for _, tt := range tests {
		shouldUpdate := tt.dirName == *goldenfile.Update

		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		if shouldUpdate {
			provider, err := NewTerraFormProvider()
			if err != nil {
				t.Fatal(err)
			}

			providerLibrary.AddProvider(terraform.AWS, provider)
			supplierLibrary.AddSupplier(NewLambdaFunctionSupplier(provider))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewLambdaFunctionDeserializer()
			client := mocks.NewMockAWSLambdaClient(tt.functionsPages)
			if tt.listError != nil {
				client = mocks.NewMockAWSLambdaErrorClient(tt.listError)
			}
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
