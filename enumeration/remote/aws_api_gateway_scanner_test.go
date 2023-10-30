package remote

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/aws"
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	resourceaws "github.com/snyk/driftctl/enumeration/resource/aws"
	"github.com/snyk/driftctl/enumeration/terraform"
	"github.com/snyk/driftctl/mocks"

	"github.com/snyk/driftctl/test/remote"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApiGatewayRestApi(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway rest apis",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllRestApis").Return([]*apigateway.RestApi{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple api gateway rest apis",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllRestApis").Return([]*apigateway.RestApi{
					{Id: awssdk.String("3of73v5ob4")},
					{Id: awssdk.String("1jitcobwol")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "3of73v5ob4")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayRestApiResourceType)

				assert.Equal(t, got[1].ResourceId(), "1jitcobwol")
				assert.Equal(t, got[1].ResourceType(), resourceaws.AwsApiGatewayRestApiResourceType)
			},
		},
		{
			test: "cannot list api gateway rest apis",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllRestApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayRestApiResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayRestApiResourceType, resourceaws.AwsApiGatewayRestApiResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayRestApiResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayRestApiEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestApiGatewayAccount(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway account",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("GetAccount").Return(nil, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "empty api gateway account",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("GetAccount").Return(&apigateway.Account{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)

				assert.Equal(t, got[0].ResourceId(), "api-gateway-account")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayAccountResourceType)
			},
		},
		{
			test: "cannot get api gateway account",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("GetAccount").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayAccountResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayAccountResourceType, resourceaws.AwsApiGatewayAccountResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayAccountResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayAccountEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestApiGatewayApiKey(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway api keys",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApiKeys").Return([]*apigateway.ApiKey{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple api gateway api keys",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApiKeys").Return([]*apigateway.ApiKey{
					{Id: awssdk.String("fuwnl8lrva")},
					{Id: awssdk.String("9ge737dd45")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "fuwnl8lrva")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayApiKeyResourceType)

				assert.Equal(t, got[1].ResourceId(), "9ge737dd45")
				assert.Equal(t, got[1].ResourceType(), resourceaws.AwsApiGatewayApiKeyResourceType)
			},
		},
		{
			test: "cannot list api gateway api keys",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApiKeys").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayApiKeyResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayApiKeyResourceType, resourceaws.AwsApiGatewayApiKeyResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayApiKeyResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayApiKeyEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestApiGatewayAuthorizer(t *testing.T) {
	dummyError := errors.New("this is an error")
	apis := []*apigateway.RestApi{
		{Id: awssdk.String("3of73v5ob4")},
		{Id: awssdk.String("1jitcobwol")},
	}

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway authorizers",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiAuthorizers", *apis[0].Id).Return([]*apigateway.Authorizer{}, nil).Once()
				repo.On("ListAllRestApiAuthorizers", *apis[1].Id).Return([]*apigateway.Authorizer{}, nil).Once()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple api gateway authorizers",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiAuthorizers", *apis[0].Id).Return([]*apigateway.Authorizer{
					{Id: awssdk.String("ypcpde")},
				}, nil).Once()
				repo.On("ListAllRestApiAuthorizers", *apis[1].Id).Return([]*apigateway.Authorizer{
					{Id: awssdk.String("bwhebj")},
				}, nil).Once()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "ypcpde")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayAuthorizerResourceType)

				assert.Equal(t, got[1].ResourceId(), "bwhebj")
				assert.Equal(t, got[1].ResourceType(), resourceaws.AwsApiGatewayAuthorizerResourceType)
			},
		},
		{
			test: "cannot list rest apis",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayAuthorizerResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayAuthorizerResourceType, resourceaws.AwsApiGatewayRestApiResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayAuthorizerResourceType, resourceaws.AwsApiGatewayRestApiResourceType),
		},
		{
			test: "cannot list api gateway resources",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiAuthorizers", *apis[0].Id).Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayAuthorizerResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayAuthorizerResourceType, resourceaws.AwsApiGatewayAuthorizerResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayAuthorizerResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayAuthorizerEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestApiGatewayStage(t *testing.T) {
	dummyError := errors.New("this is an error")
	apis := []*apigateway.RestApi{
		{Id: awssdk.String("3of73v5ob4")},
	}

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway stages",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiStages", *apis[0].Id).Return([]*apigateway.Stage{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple api gateway stages",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiStages", *apis[0].Id).Return([]*apigateway.Stage{
					{StageName: awssdk.String("foo")},
					{StageName: awssdk.String("baz")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "ags-3of73v5ob4-foo")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayStageResourceType)

				assert.Equal(t, got[1].ResourceId(), "ags-3of73v5ob4-baz")
				assert.Equal(t, got[1].ResourceType(), resourceaws.AwsApiGatewayStageResourceType)
			},
		},
		{
			test: "cannot list rest apis",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayStageResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayStageResourceType, resourceaws.AwsApiGatewayRestApiResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayStageResourceType, resourceaws.AwsApiGatewayRestApiResourceType),
		},
		{
			test: "cannot list api gateway stages",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiStages", *apis[0].Id).Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayStageResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayStageResourceType, resourceaws.AwsApiGatewayStageResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayStageResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayStageEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestApiGatewayResource(t *testing.T) {
	dummyError := errors.New("this is an error")
	apis := []*apigateway.RestApi{
		{Id: awssdk.String("3of73v5ob4")},
	}

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway resources",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiResources", *apis[0].Id).Return([]*apigateway.Resource{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple api gateway resources",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiResources", *apis[0].Id).Return([]*apigateway.Resource{
					{Id: awssdk.String("21zk4y"), Path: awssdk.String("/")},
					{Id: awssdk.String("2ltv32p058"), Path: awssdk.String("/")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "21zk4y")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayResourceResourceType)

				assert.Equal(t, got[1].ResourceId(), "2ltv32p058")
				assert.Equal(t, got[1].ResourceType(), resourceaws.AwsApiGatewayResourceResourceType)
			},
		},
		{
			test: "cannot list rest apis",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayResourceResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayResourceResourceType, resourceaws.AwsApiGatewayRestApiResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayResourceResourceType, resourceaws.AwsApiGatewayRestApiResourceType),
		},
		{
			test: "cannot list api gateway resources",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiResources", *apis[0].Id).Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayResourceResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayResourceResourceType, resourceaws.AwsApiGatewayResourceResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayResourceResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayResourceEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestApiGatewayDomainName(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway domain names",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllDomainNames").Return([]*apigateway.DomainName{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "single api gateway domain name",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllDomainNames").Return([]*apigateway.DomainName{
					{DomainName: awssdk.String("example-driftctl.com")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)

				assert.Equal(t, got[0].ResourceId(), "example-driftctl.com")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayDomainNameResourceType)
			},
		},
		{
			test: "cannot list api gateway domain names",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllDomainNames").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayDomainNameResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayDomainNameResourceType, resourceaws.AwsApiGatewayDomainNameResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayDomainNameResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayDomainNameEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestApiGatewayVpcLink(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway vpc links",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllVpcLinks").Return([]*apigateway.UpdateVpcLinkOutput{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "single api gateway vpc link",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllVpcLinks").Return([]*apigateway.UpdateVpcLinkOutput{
					{Id: awssdk.String("ipu24n")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)

				assert.Equal(t, got[0].ResourceId(), "ipu24n")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayVpcLinkResourceType)
			},
		},
		{
			test: "cannot list api gateway vpc links",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllVpcLinks").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayVpcLinkResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayVpcLinkResourceType, resourceaws.AwsApiGatewayVpcLinkResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayVpcLinkResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayVpcLinkEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestApiGatewayRequestValidator(t *testing.T) {
	dummyError := errors.New("this is an error")
	apis := []*apigateway.RestApi{
		{Id: awssdk.String("vryjzimtj1")},
	}

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway request validators",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiRequestValidators", *apis[0].Id).Return([]*apigateway.UpdateRequestValidatorOutput{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple api gateway request validators",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiRequestValidators", *apis[0].Id).Return([]*apigateway.UpdateRequestValidatorOutput{
					{Id: awssdk.String("ywlcuf")},
					{Id: awssdk.String("qmpbs8")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "ywlcuf")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayRequestValidatorResourceType)

				assert.Equal(t, got[1].ResourceId(), "qmpbs8")
				assert.Equal(t, got[1].ResourceType(), resourceaws.AwsApiGatewayRequestValidatorResourceType)
			},
		},
		{
			test: "cannot list rest apis",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayRequestValidatorResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayRequestValidatorResourceType, resourceaws.AwsApiGatewayRestApiResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayRequestValidatorResourceType, resourceaws.AwsApiGatewayRestApiResourceType),
		},
		{
			test: "cannot list api gateway request validators",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiRequestValidators", *apis[0].Id).Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayRequestValidatorResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayRequestValidatorResourceType, resourceaws.AwsApiGatewayRequestValidatorResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayRequestValidatorResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayRequestValidatorEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestApiGatewayRestApiPolicy(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway rest api policies",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllRestApis").Return([]*apigateway.RestApi{
					{Id: awssdk.String("3of73v5ob4")},
					{Id: awssdk.String("9x7kq9pbyh"), Policy: awssdk.String("")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple api gateway rest api policies",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllRestApis").Return([]*apigateway.RestApi{
					{Id: awssdk.String("c3n3aqga5d"), Policy: awssdk.String("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":\"*\"},\"Action\":\"execute-api:Invoke\",\"Resource\":\"arn:aws:execute-api:us-east-1:111111111111:c3n3aqga5d/*\",\"Condition\":{\"IpAddress\":{\"aws:SourceIp\":\"123.123.123.123/32\"}}}]}")},
					{Id: awssdk.String("9y1eus3hr7"), Policy: awssdk.String("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":\"*\"},\"Action\":\"execute-api:Invoke\",\"Resource\":\"arn:aws:execute-api:us-east-1:111111111111:9y1eus3hr7/*\",\"Condition\":{\"IpAddress\":{\"aws:SourceIp\":\"123.123.123.123/32\"}}}]}")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "c3n3aqga5d")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayRestApiPolicyResourceType)

				assert.Equal(t, got[1].ResourceId(), "9y1eus3hr7")
				assert.Equal(t, got[1].ResourceType(), resourceaws.AwsApiGatewayRestApiPolicyResourceType)
			},
		},
		{
			test: "cannot list rest apis",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayRestApiPolicyResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayRestApiPolicyResourceType, resourceaws.AwsApiGatewayRestApiResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayRestApiPolicyResourceType, resourceaws.AwsApiGatewayRestApiResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayRestApiPolicyEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestApiGatewayBasePathMapping(t *testing.T) {
	dummyError := errors.New("this is an error")
	domainNames := []*apigateway.DomainName{
		{DomainName: awssdk.String("example-driftctl.com")},
	}

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no domain name base path mappings",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllDomainNames").Return(domainNames, nil)
				repo.On("ListAllDomainNameBasePathMappings", *domainNames[0].DomainName).Return([]*apigateway.BasePathMapping{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple domain name base path mappings",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllDomainNames").Return(domainNames, nil)
				repo.On("ListAllDomainNameBasePathMappings", *domainNames[0].DomainName).Return([]*apigateway.BasePathMapping{
					{BasePath: awssdk.String("foo")},
					{BasePath: awssdk.String("(none)")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "example-driftctl.com/foo")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayBasePathMappingResourceType)

				assert.Equal(t, got[1].ResourceId(), "example-driftctl.com/")
				assert.Equal(t, got[1].ResourceType(), resourceaws.AwsApiGatewayBasePathMappingResourceType)
			},
		},
		{
			test: "cannot list domain names",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllDomainNames").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayBasePathMappingResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayBasePathMappingResourceType, resourceaws.AwsApiGatewayDomainNameResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayBasePathMappingResourceType, resourceaws.AwsApiGatewayDomainNameResourceType),
		},
		{
			test: "cannot list domain name base path mappings",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllDomainNames").Return(domainNames, nil)
				repo.On("ListAllDomainNameBasePathMappings", *domainNames[0].DomainName).Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayBasePathMappingResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayBasePathMappingResourceType, resourceaws.AwsApiGatewayBasePathMappingResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayBasePathMappingResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayBasePathMappingEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestApiGatewayMethod(t *testing.T) {
	dummyError := errors.New("this is an error")
	apis := []*apigateway.RestApi{
		{Id: awssdk.String("vryjzimtj1")},
	}

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway methods",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiResources", *apis[0].Id).Return([]*apigateway.Resource{
					{Id: awssdk.String("hl7ksq"), Path: awssdk.String("/foo")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple api gateway methods",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiResources", *apis[0].Id).Return([]*apigateway.Resource{
					{Id: awssdk.String("hl7ksq"), Path: awssdk.String("/foo"), ResourceMethods: map[string]*apigateway.Method{
						"GET":    {},
						"POST":   {},
						"DELETE": {},
					}},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 3)

				assert.Equal(t, got[0].ResourceId(), "agm-vryjzimtj1-hl7ksq-DELETE")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayMethodResourceType)

				assert.Equal(t, got[1].ResourceId(), "agm-vryjzimtj1-hl7ksq-GET")
				assert.Equal(t, got[1].ResourceType(), resourceaws.AwsApiGatewayMethodResourceType)

				assert.Equal(t, got[2].ResourceId(), "agm-vryjzimtj1-hl7ksq-POST")
				assert.Equal(t, got[2].ResourceType(), resourceaws.AwsApiGatewayMethodResourceType)
			},
		},
		{
			test: "cannot list rest apis",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayMethodResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayMethodResourceType, resourceaws.AwsApiGatewayRestApiResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayMethodResourceType, resourceaws.AwsApiGatewayRestApiResourceType),
		},
		{
			test: "cannot list api gateway resources",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiResources", *apis[0].Id).Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayMethodResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayMethodResourceType, resourceaws.AwsApiGatewayResourceResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayMethodResourceType, resourceaws.AwsApiGatewayResourceResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayMethodEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := remote.NewSortableScanner(NewScanner(remoteLibrary, alerter, testFilter))
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestApiGatewayModel(t *testing.T) {
	dummyError := errors.New("this is an error")
	apis := []*apigateway.RestApi{
		{Id: awssdk.String("vryjzimtj1")},
	}

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway models",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiModels", *apis[0].Id).Return([]*apigateway.Model{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple api gateway models",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiModels", *apis[0].Id).Return([]*apigateway.Model{
					{Id: awssdk.String("g68a4s")},
					{Id: awssdk.String("85v536")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "g68a4s")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayModelResourceType)

				assert.Equal(t, got[1].ResourceId(), "85v536")
				assert.Equal(t, got[1].ResourceType(), resourceaws.AwsApiGatewayModelResourceType)
			},
		},
		{
			test: "cannot list rest apis",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayModelResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayModelResourceType, resourceaws.AwsApiGatewayRestApiResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayModelResourceType, resourceaws.AwsApiGatewayRestApiResourceType),
		},
		{
			test: "cannot list api gateway models",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiModels", *apis[0].Id).Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayModelResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayModelResourceType, resourceaws.AwsApiGatewayModelResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayModelResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayModelEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestApiGatewayMethodResponse(t *testing.T) {
	dummyError := errors.New("this is an error")
	apis := []*apigateway.RestApi{
		{Id: awssdk.String("vryjzimtj1")},
	}

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway method responses",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiResources", *apis[0].Id).Return([]*apigateway.Resource{
					{Id: awssdk.String("hl7ksq"), Path: awssdk.String("/foo"), ResourceMethods: map[string]*apigateway.Method{
						"GET": {},
					}},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple api gateway method responses",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiResources", *apis[0].Id).Return([]*apigateway.Resource{
					{Id: awssdk.String("hl7ksq"), Path: awssdk.String("/foo"), ResourceMethods: map[string]*apigateway.Method{
						"GET": {MethodResponses: map[string]*apigateway.MethodResponse{
							"200": {},
							"404": {},
							"503": {},
						}},
					}},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 3)

				assert.Equal(t, got[0].ResourceId(), "agmr-vryjzimtj1-hl7ksq-GET-200")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayMethodResponseResourceType)

				assert.Equal(t, got[1].ResourceId(), "agmr-vryjzimtj1-hl7ksq-GET-404")
				assert.Equal(t, got[1].ResourceType(), resourceaws.AwsApiGatewayMethodResponseResourceType)

				assert.Equal(t, got[2].ResourceId(), "agmr-vryjzimtj1-hl7ksq-GET-503")
				assert.Equal(t, got[2].ResourceType(), resourceaws.AwsApiGatewayMethodResponseResourceType)
			},
		},
		{
			test: "cannot list rest apis",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayMethodResponseResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayMethodResponseResourceType, resourceaws.AwsApiGatewayRestApiResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayMethodResponseResourceType, resourceaws.AwsApiGatewayRestApiResourceType),
		},
		{
			test: "cannot list api gateway resources",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiResources", *apis[0].Id).Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayMethodResponseResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayMethodResponseResourceType, resourceaws.AwsApiGatewayResourceResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayMethodResponseResourceType, resourceaws.AwsApiGatewayResourceResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayMethodResponseEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := remote.NewSortableScanner(NewScanner(remoteLibrary, alerter, testFilter))
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestApiGatewayGatewayResponse(t *testing.T) {
	dummyError := errors.New("this is an error")
	apis := []*apigateway.RestApi{
		{Id: awssdk.String("vryjzimtj1")},
	}

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway gateway responses",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiGatewayResponses", *apis[0].Id).Return([]*apigateway.UpdateGatewayResponseOutput{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple api gateway gateway responses",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiGatewayResponses", *apis[0].Id).Return([]*apigateway.UpdateGatewayResponseOutput{
					{ResponseType: awssdk.String("UNAUTHORIZED")},
					{ResponseType: awssdk.String("ACCESS_DENIED")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "aggr-vryjzimtj1-UNAUTHORIZED")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayGatewayResponseResourceType)

				assert.Equal(t, got[1].ResourceId(), "aggr-vryjzimtj1-ACCESS_DENIED")
				assert.Equal(t, got[1].ResourceType(), resourceaws.AwsApiGatewayGatewayResponseResourceType)
			},
		},
		{
			test: "cannot list rest apis",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayGatewayResponseResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayGatewayResponseResourceType, resourceaws.AwsApiGatewayRestApiResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayGatewayResponseResourceType, resourceaws.AwsApiGatewayRestApiResourceType),
		},
		{
			test: "cannot list api gateway gateway responses",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiGatewayResponses", *apis[0].Id).Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayGatewayResponseResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayGatewayResponseResourceType, resourceaws.AwsApiGatewayGatewayResponseResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayGatewayResponseResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayGatewayResponseEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestApiGatewayMethodSettings(t *testing.T) {
	dummyError := errors.New("this is an error")
	apis := []*apigateway.RestApi{
		{Id: awssdk.String("vryjzimtj1")},
	}

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway method settings",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiStages", *apis[0].Id).Return([]*apigateway.Stage{
					{StageName: awssdk.String("foo"), MethodSettings: map[string]*apigateway.MethodSetting{}},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple api gateway method settings",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiStages", *apis[0].Id).Return([]*apigateway.Stage{
					{StageName: awssdk.String("foo"), MethodSettings: map[string]*apigateway.MethodSetting{
						"*/*":        {},
						"foo/GET":    {},
						"foo/DELETE": {},
					}},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 3)

				assert.Equal(t, got[0].ResourceId(), "vryjzimtj1-foo-*/*")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayMethodSettingsResourceType)

				assert.Equal(t, got[1].ResourceId(), "vryjzimtj1-foo-foo/DELETE")
				assert.Equal(t, got[1].ResourceType(), resourceaws.AwsApiGatewayMethodSettingsResourceType)

				assert.Equal(t, got[2].ResourceId(), "vryjzimtj1-foo-foo/GET")
				assert.Equal(t, got[2].ResourceType(), resourceaws.AwsApiGatewayMethodSettingsResourceType)
			},
		},
		{
			test: "cannot list rest apis",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayMethodSettingsResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayMethodSettingsResourceType, resourceaws.AwsApiGatewayRestApiResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayMethodSettingsResourceType, resourceaws.AwsApiGatewayRestApiResourceType),
		},
		{
			test: "cannot list api gateway settings",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiStages", *apis[0].Id).Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayMethodSettingsResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayMethodSettingsResourceType, resourceaws.AwsApiGatewayStageResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayMethodSettingsResourceType, resourceaws.AwsApiGatewayStageResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayMethodSettingsEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := remote.NewSortableScanner(NewScanner(remoteLibrary, alerter, testFilter))
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestApiGatewayIntegration(t *testing.T) {
	dummyError := errors.New("this is an error")
	apis := []*apigateway.RestApi{
		{Id: awssdk.String("u7jce3lokk")},
	}

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway integrations",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiResources", *apis[0].Id).Return([]*apigateway.Resource{
					{Id: awssdk.String("z9ag20"), Path: awssdk.String("/foo")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple api gateway integrations",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiResources", *apis[0].Id).Return([]*apigateway.Resource{
					{Id: awssdk.String("z9ag20"), Path: awssdk.String("/foo"), ResourceMethods: map[string]*apigateway.Method{
						"GET":    {},
						"POST":   {},
						"DELETE": {},
					}},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 3)

				assert.Equal(t, got[0].ResourceId(), "agi-u7jce3lokk-z9ag20-DELETE")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayIntegrationResourceType)

				assert.Equal(t, got[1].ResourceId(), "agi-u7jce3lokk-z9ag20-GET")
				assert.Equal(t, got[1].ResourceType(), resourceaws.AwsApiGatewayIntegrationResourceType)

				assert.Equal(t, got[2].ResourceId(), "agi-u7jce3lokk-z9ag20-POST")
				assert.Equal(t, got[2].ResourceType(), resourceaws.AwsApiGatewayIntegrationResourceType)
			},
		},
		{
			test: "cannot list rest apis",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayIntegrationResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayIntegrationResourceType, resourceaws.AwsApiGatewayRestApiResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayIntegrationResourceType, resourceaws.AwsApiGatewayRestApiResourceType),
		},
		{
			test: "cannot list api gateway resources",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiResources", *apis[0].Id).Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayIntegrationResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayIntegrationResourceType, resourceaws.AwsApiGatewayResourceResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayIntegrationResourceType, resourceaws.AwsApiGatewayResourceResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayIntegrationEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := remote.NewSortableScanner(NewScanner(remoteLibrary, alerter, testFilter))
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestApiGatewayIntegrationResponse(t *testing.T) {
	dummyError := errors.New("this is an error")
	apis := []*apigateway.RestApi{
		{Id: awssdk.String("u7jce3lokk")},
	}

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway integration responses",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiResources", *apis[0].Id).Return([]*apigateway.Resource{
					{Id: awssdk.String("z9ag20"), Path: awssdk.String("/foo"), ResourceMethods: map[string]*apigateway.Method{
						"GET": {},
					}},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple api gateway integration responses",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiResources", *apis[0].Id).Return([]*apigateway.Resource{
					{Id: awssdk.String("z9ag20"), Path: awssdk.String("/foo"), ResourceMethods: map[string]*apigateway.Method{
						"GET": {
							MethodIntegration: &apigateway.Integration{
								IntegrationResponses: map[string]*apigateway.IntegrationResponse{
									"200": {},
									"302": {},
								},
							},
						},
					}},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "agir-u7jce3lokk-z9ag20-GET-200")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayIntegrationResponseResourceType)

				assert.Equal(t, got[1].ResourceId(), "agir-u7jce3lokk-z9ag20-GET-302")
				assert.Equal(t, got[1].ResourceType(), resourceaws.AwsApiGatewayIntegrationResponseResourceType)
			},
		},
		{
			test: "cannot list rest apis",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayIntegrationResponseResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayIntegrationResponseResourceType, resourceaws.AwsApiGatewayRestApiResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayIntegrationResponseResourceType, resourceaws.AwsApiGatewayRestApiResourceType),
		},
		{
			test: "cannot list api gateway resources",
			mocks: func(repo *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllRestApis").Return(apis, nil)
				repo.On("ListAllRestApiResources", *apis[0].Id).Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayIntegrationResponseResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayIntegrationResponseResourceType, resourceaws.AwsApiGatewayResourceResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayIntegrationResponseResourceType, resourceaws.AwsApiGatewayResourceResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayIntegrationResponseEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := remote.NewSortableScanner(NewScanner(remoteLibrary, alerter, testFilter))
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}
