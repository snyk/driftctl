package aws

import (
	"context"
	"testing"

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
	}
	for _, tt := range tests {
		shouldUpdate := tt.dirName == *goldenfile.Update
		if shouldUpdate {
			provider, err := NewTerraFormProvider()
			if err != nil {
				t.Fatal(err)
			}

			terraform.AddProvider(terraform.AWS, provider)
			resource.AddSupplier(NewLambdaFunctionSupplier(provider.Runner(), lambda.New(provider.session)))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, terraform.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewLambdaFunctionDeserializer()
			s := &LambdaFunctionSupplier{
				provider,
				deserializer,
				mocks.NewMockAWSLambdaClient(tt.functionsPages),
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			if tt.err != err {
				t.Errorf("Expected error %+v got %+v", tt.err, err)
			}

			test.CtyTestDiff(got, tt.dirName, provider, deserializer, shouldUpdate, t)
		})
	}
}
