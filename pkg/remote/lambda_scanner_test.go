package remote

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/remote/aws"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	terraform2 "github.com/cloudskiff/driftctl/test/terraform"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	"github.com/stretchr/testify/assert"
)

func TestScanLambdaFunction(t *testing.T) {

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
						FunctionName: awssdk.String("foo"),
					},
					{
						FunctionName: awssdk.String("bar"),
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
						FunctionName: awssdk.String("foo"),
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
			err: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			session := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			alerter := &mocks.AlerterInterface{}
			alerter.On("SendAlert", mock.Anything, mock.Anything).Maybe().Return()

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			fakeRepo := &repository.MockLambdaRepository{}
			c.mocks(fakeRepo)
			var repo repository.LambdaRepository = fakeRepo
			providerVersion := "3.19.0"
			realProvider, err := terraform2.InitTestAwsProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				repo = repository.NewLambdaRepository(session, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewLambdaFunctionEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsLambdaFunctionResourceType, aws.NewLambdaFunctionDetailsFetcher(provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsLambdaFunctionResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestScanLambdaEventSourceMapping(t *testing.T) {

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
						UUID: awssdk.String("13ff66f8-37eb-4ad6-a0a8-594fea72df4f"),
					},
					{
						UUID: awssdk.String("4ad7e2b3-79e9-4713-9d9d-5af2c01d9058"),
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
						UUID: awssdk.String("1aa9c4a0-060b-41c1-a9ae-dc304ebcdb00"),
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
			err: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			session := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			alerter := &mocks.AlerterInterface{}
			alerter.On("SendAlert", mock.Anything, mock.Anything).Maybe().Return()

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			fakeRepo := &repository.MockLambdaRepository{}
			c.mocks(fakeRepo)
			var repo repository.LambdaRepository = fakeRepo
			providerVersion := "3.19.0"
			realProvider, err := terraform2.InitTestAwsProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				repo = repository.NewLambdaRepository(session, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewLambdaEventSourceMappingEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsLambdaEventSourceMappingResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsLambdaEventSourceMappingResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsLambdaEventSourceMappingResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}
