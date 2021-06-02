package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	testresource "github.com/cloudskiff/driftctl/test/resource"

	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	"github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
)

func TestEventLambdaSourceMappingSupplier_Resources(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repo *repository.MockLambdaRepository)
		err     error
	}{
		{
			test:    "no EventSourceMapping",
			dirName: "lambda_source_mapping_empty",
			mocks: func(repo *repository.MockLambdaRepository) {
				repo.On("ListAllLambdaEventSourceMappings").Return([]*lambda.EventSourceMappingConfiguration{}, nil)
			},
			err: nil,
		},
		{
			test:    "with 2 sqs EventSourceMapping",
			dirName: "lambda_source_mapping_sqs_multiple",
			mocks: func(repo *repository.MockLambdaRepository) {
				repo.On("ListAllLambdaEventSourceMappings").Return([]*lambda.EventSourceMappingConfiguration{
					{
						UUID: aws.String("13ff66f8-37eb-4ad6-a0a8-594fea72df4f"),
					},
					{
						UUID: aws.String("4ad7e2b3-79e9-4713-9d9d-5af2c01d9058"),
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "with dynamo EventSourceMapping",
			dirName: "lambda_source_mapping_dynamo_multiple",
			mocks: func(repo *repository.MockLambdaRepository) {
				repo.On("ListAllLambdaEventSourceMappings").Return([]*lambda.EventSourceMappingConfiguration{
					{
						UUID: aws.String("1aa9c4a0-060b-41c1-a9ae-dc304ebcdb00"),
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list lambda functions",
			dirName: "lambda_function_empty",
			mocks: func(repo *repository.MockLambdaRepository) {
				repo.On("ListAllLambdaEventSourceMappings").Return([]*lambda.EventSourceMappingConfiguration{}, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsLambdaEventSourceMappingResourceType),
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
			supplierLibrary.AddSupplier(NewLambdaEventSourceMappingSupplier(provider, deserializer))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			client := &repository.MockLambdaRepository{}
			tt.mocks(client)
			s := &LambdaEventSourceMappingSupplier{
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
