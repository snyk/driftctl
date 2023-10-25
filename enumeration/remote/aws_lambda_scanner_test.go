package remote

import (
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/aws"
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/terraform"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/mocks"

	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/mock"

	"github.com/snyk/driftctl/enumeration/resource"
	resourceaws "github.com/snyk/driftctl/enumeration/resource/aws"

	"github.com/snyk/driftctl/test/goldenfile"
	"github.com/stretchr/testify/assert"
)

func TestScanLambdaFunction(t *testing.T) {
	tests := []struct {
		test           string
		dirName        string
		mocks          func(*repository.MockLambdaRepository, *mocks.AlerterInterface)
		assertExpected func(*testing.T, []*resource.Resource)
		err            error
	}{
		{
			test:    "no lambda functions",
			dirName: "aws_lambda_function_empty",
			mocks: func(repo *repository.MockLambdaRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllLambdaFunctions").Return([]*lambda.FunctionConfiguration{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			err: nil,
		},
		{
			test:    "with lambda functions",
			dirName: "aws_lambda_function_multiple",
			mocks: func(repo *repository.MockLambdaRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllLambdaFunctions").Return([]*lambda.FunctionConfiguration{
					{
						FunctionName: awssdk.String("foo"),
					},
					{
						FunctionName: awssdk.String("bar"),
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, "foo", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsLambdaFunctionResourceType, got[0].ResourceType())

				assert.Equal(t, "bar", got[1].ResourceId())
				assert.Equal(t, resourceaws.AwsLambdaFunctionResourceType, got[1].ResourceType())
			},
			err: nil,
		},
		{
			test:    "One lambda with signing",
			dirName: "aws_lambda_function_signed",
			mocks: func(repo *repository.MockLambdaRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllLambdaFunctions").Return([]*lambda.FunctionConfiguration{
					{
						FunctionName: awssdk.String("foo"),
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)

				assert.Equal(t, "foo", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsLambdaFunctionResourceType, got[0].ResourceType())
			},
			err: nil,
		},
		{
			test:    "cannot list lambda functions",
			dirName: "aws_lambda_function_empty",
			mocks: func(repo *repository.MockLambdaRepository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repo.On("ListAllLambdaFunctions").Return([]*lambda.FunctionConfiguration{}, awsError)

				alerter.On("SendAlert", resourceaws.AwsLambdaFunctionResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsLambdaFunctionResourceType, resourceaws.AwsLambdaFunctionResourceType), alerts.EnumerationPhase)).Return()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			err: nil,
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			session := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockLambdaRepository{}
			c.mocks(fakeRepo, alerter)

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

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestScanLambdaEventSourceMapping(t *testing.T) {
	tests := []struct {
		test           string
		dirName        string
		mocks          func(*repository.MockLambdaRepository, *mocks.AlerterInterface)
		assertExpected func(*testing.T, []*resource.Resource)
		err            error
	}{
		{
			test:    "no EventSourceMapping",
			dirName: "aws_lambda_source_mapping_empty",
			mocks: func(repo *repository.MockLambdaRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllLambdaEventSourceMappings").Return([]*lambda.EventSourceMappingConfiguration{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			err: nil,
		},
		{
			test:    "with 2 sqs EventSourceMapping",
			dirName: "aws_lambda_source_mapping_sqs_multiple",
			mocks: func(repo *repository.MockLambdaRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllLambdaEventSourceMappings").Return([]*lambda.EventSourceMappingConfiguration{
					{
						UUID: awssdk.String("13ff66f8-37eb-4ad6-a0a8-594fea72df4f"),
					},
					{
						UUID: awssdk.String("4ad7e2b3-79e9-4713-9d9d-5af2c01d9058"),
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, "13ff66f8-37eb-4ad6-a0a8-594fea72df4f", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsLambdaEventSourceMappingResourceType, got[0].ResourceType())

				assert.Equal(t, "4ad7e2b3-79e9-4713-9d9d-5af2c01d9058", got[1].ResourceId())
				assert.Equal(t, resourceaws.AwsLambdaEventSourceMappingResourceType, got[1].ResourceType())
			},
			err: nil,
		},
		{
			test:    "with dynamo EventSourceMapping",
			dirName: "aws_lambda_source_mapping_dynamo_multiple",
			mocks: func(repo *repository.MockLambdaRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllLambdaEventSourceMappings").Return([]*lambda.EventSourceMappingConfiguration{
					{
						UUID: awssdk.String("1aa9c4a0-060b-41c1-a9ae-dc304ebcdb00"),
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)

				assert.Equal(t, "1aa9c4a0-060b-41c1-a9ae-dc304ebcdb00", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsLambdaEventSourceMappingResourceType, got[0].ResourceType())
			},
			err: nil,
		},
		{
			test:    "cannot list lambda functions",
			dirName: "aws_lambda_function_empty",
			mocks: func(repo *repository.MockLambdaRepository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repo.On("ListAllLambdaEventSourceMappings").Return([]*lambda.EventSourceMappingConfiguration{}, awsError)

				alerter.On("SendAlert", resourceaws.AwsLambdaEventSourceMappingResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsLambdaEventSourceMappingResourceType, resourceaws.AwsLambdaEventSourceMappingResourceType), alerts.EnumerationPhase)).Return()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			err: nil,
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			session := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockLambdaRepository{}
			c.mocks(fakeRepo, alerter)

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

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}
