package remote

import (
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/aws"
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/terraform"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration/resource"
	resourceaws "github.com/snyk/driftctl/enumeration/resource/aws"
	"github.com/snyk/driftctl/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApiGatewayV2Api(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayV2Repository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway v2 api",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return([]*apigatewayv2.Api{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "single api gateway v2 api",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return([]*apigatewayv2.Api{
					{ApiId: awssdk.String("f5vdrg12tk")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)

				assert.Equal(t, got[0].ResourceId(), "f5vdrg12tk")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayV2ApiResourceType)
			},
		},
		{
			test: "cannot list api gateway v2 apis",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2ApiResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2ApiResourceType, resourceaws.AwsApiGatewayV2ApiResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayV2ApiResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayV2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayV2Repository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayV2ApiEnumerator(repo, factory))

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

func TestApiGatewayV2Route(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayV2Repository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway v2 api",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return([]*apigatewayv2.Api{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "single api gateway v2 api with a single route",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return([]*apigatewayv2.Api{
					{ApiId: awssdk.String("a-gateway")},
				}, nil)
				repository.On("ListAllApiRoutes", awssdk.String("a-gateway")).
					Return([]*apigatewayv2.Route{{
						RouteId:  awssdk.String("a-route"),
						RouteKey: awssdk.String("POST /an-example"),
					}}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)

				assert.Equal(t, "a-route", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsApiGatewayV2RouteResourceType, got[0].ResourceType())
				expectedAttrs := &resource.Attributes{
					"api_id":    "a-gateway",
					"route_key": "POST /an-example",
				}
				assert.Equal(t, expectedAttrs, got[0].Attributes())
			},
		},
		{
			test: "cannot list api gateway v2 apis",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2RouteResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2ApiResourceType, resourceaws.AwsApiGatewayV2RouteResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2RouteResourceType, resourceaws.AwsApiGatewayV2ApiResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayV2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayV2Repository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayV2RouteEnumerator(repo, factory))

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

func TestApiGatewayV2Deployment(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayV2Repository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "single api gateway v2 api with a single deployment",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return([]*apigatewayv2.Api{
					{ApiId: awssdk.String("a-gateway")},
				}, nil)
				repository.On("ListAllApiDeployments", awssdk.String("a-gateway")).
					Return([]*apigatewayv2.Deployment{{
						DeploymentId: awssdk.String("a-deployment"),
					}}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)

				assert.Equal(t, "a-deployment", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsApiGatewayV2DeploymentResourceType, got[0].ResourceType())
				expectedAttrs := &resource.Attributes{}
				assert.Equal(t, expectedAttrs, got[0].Attributes())
			},
		},
		{
			test: "no API gateways",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return([]*apigatewayv2.Api{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "single API gateway with no deployments",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return([]*apigatewayv2.Api{
					{ApiId: awssdk.String("a-gateway")},
				}, nil)
				repository.On("ListAllApiDeployments", awssdk.String("a-gateway")).
					Return([]*apigatewayv2.Deployment{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "error listing API gateways",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2DeploymentResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2ApiResourceType, resourceaws.AwsApiGatewayV2RouteResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2DeploymentResourceType, resourceaws.AwsApiGatewayV2ApiResourceType),
		},
		{
			test: "error listing deployments of an API",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return([]*apigatewayv2.Api{
					{ApiId: awssdk.String("a-gateway")},
				}, nil)
				repository.On("ListAllApiDeployments", awssdk.String("a-gateway")).Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2DeploymentResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2DeploymentResourceType, resourceaws.AwsApiGatewayV2DeploymentResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2DeploymentResourceType, resourceaws.AwsApiGatewayV2DeploymentResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayV2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayV2Repository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayV2DeploymentEnumerator(repo, factory))

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

func TestApiGatewayV2VpcLink(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayV2Repository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway v2 vpc links",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllVpcLinks").Return([]*apigatewayv2.VpcLink{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "single api gateway v2 vpc link",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllVpcLinks").Return([]*apigatewayv2.VpcLink{
					{VpcLinkId: awssdk.String("b8r351")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)

				assert.Equal(t, got[0].ResourceId(), "b8r351")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayV2VpcLinkResourceType)
			},
		},
		{
			test: "cannot list api gateway v2 vpc links",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllVpcLinks").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2VpcLinkResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2VpcLinkResourceType, resourceaws.AwsApiGatewayV2VpcLinkResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayV2VpcLinkResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayV2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayV2Repository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayV2VpcLinkEnumerator(repo, factory))

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

func TestApiGatewayV2Authorizer(t *testing.T) {
	dummyError := errors.New("this is an error")

	apis := []*apigatewayv2.Api{
		{ApiId: awssdk.String("bmyl5c6huh")},
		{ApiId: awssdk.String("blghshbgte")},
	}

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayV2Repository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway v2 authorizers",
			mocks: func(repo *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllApis").Return(apis, nil)
				repo.On("ListAllApiAuthorizers", *apis[0].ApiId).Return([]*apigatewayv2.Authorizer{}, nil).Once()
				repo.On("ListAllApiAuthorizers", *apis[1].ApiId).Return([]*apigatewayv2.Authorizer{}, nil).Once()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple api gateway v2 authorizers",
			mocks: func(repo *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllApis").Return(apis, nil)
				repo.On("ListAllApiAuthorizers", *apis[0].ApiId).Return([]*apigatewayv2.Authorizer{
					{AuthorizerId: awssdk.String("xaappu")},
				}, nil).Once()
				repo.On("ListAllApiAuthorizers", *apis[1].ApiId).Return([]*apigatewayv2.Authorizer{
					{AuthorizerId: awssdk.String("bwhebj")},
				}, nil).Once()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "xaappu")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayV2AuthorizerResourceType)

				assert.Equal(t, got[1].ResourceId(), "bwhebj")
				assert.Equal(t, got[1].ResourceType(), resourceaws.AwsApiGatewayV2AuthorizerResourceType)
			},
		},
		{
			test: "cannot list apis",
			mocks: func(repo *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2AuthorizerResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2AuthorizerResourceType, resourceaws.AwsApiGatewayV2ApiResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2AuthorizerResourceType, resourceaws.AwsApiGatewayV2ApiResourceType),
		},
		{
			test: "cannot list api gateway v2 authorizers",
			mocks: func(repo *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllApis").Return(apis, nil)
				repo.On("ListAllApiAuthorizers", *apis[0].ApiId).Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2AuthorizerResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2AuthorizerResourceType, resourceaws.AwsApiGatewayV2AuthorizerResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayV2AuthorizerResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayV2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayV2Repository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayV2AuthorizerEnumerator(repo, factory))

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

func TestApiGatewayV2Integration(t *testing.T) {
	dummyError := errors.New("this is an error")

	apis := []*apigatewayv2.Api{
		{ApiId: awssdk.String("bmyl5c6huh")},
		{ApiId: awssdk.String("blghshbgte")},
	}

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayV2Repository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway v2 integrations",
			mocks: func(repo *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllApis").Return(apis, nil)
				repo.On("ListAllApiIntegrations", *apis[0].ApiId).Return([]*apigatewayv2.Integration{}, nil).Once()
				repo.On("ListAllApiIntegrations", *apis[1].ApiId).Return([]*apigatewayv2.Integration{}, nil).Once()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple api gateway v2 integrations",
			mocks: func(repo *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllApis").Return(apis, nil)
				repo.On("ListAllApiIntegrations", *apis[0].ApiId).Return([]*apigatewayv2.Integration{
					{
						IntegrationId:   awssdk.String("xaappu"),
						IntegrationType: awssdk.String("MOCK"),
					},
				}, nil).Once()
				repo.On("ListAllApiIntegrations", *apis[1].ApiId).Return([]*apigatewayv2.Integration{
					{
						IntegrationId:   awssdk.String("bwhebj"),
						IntegrationType: awssdk.String("MOCK"),
					},
				}, nil).Once()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "xaappu")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayV2IntegrationResourceType)

				assert.Equal(t, got[1].ResourceId(), "bwhebj")
				assert.Equal(t, got[1].ResourceType(), resourceaws.AwsApiGatewayV2IntegrationResourceType)
			},
		},
		{
			test: "cannot list apis",
			mocks: func(repo *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2IntegrationResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2IntegrationResourceType, resourceaws.AwsApiGatewayV2ApiResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2IntegrationResourceType, resourceaws.AwsApiGatewayV2ApiResourceType),
		},
		{
			test: "cannot list api gateway v2 integrations",
			mocks: func(repo *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllApis").Return(apis, nil)
				repo.On("ListAllApiIntegrations", *apis[0].ApiId).Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2IntegrationResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2IntegrationResourceType, resourceaws.AwsApiGatewayV2IntegrationResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayV2IntegrationResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayV2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayV2Repository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayV2IntegrationEnumerator(repo, factory))

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

func TestApiGatewayV2Model(t *testing.T) {
	dummyError := errors.New("this is an error")

	apis := []*apigatewayv2.Api{
		{ApiId: awssdk.String("bmyl5c6huh")},
		{ApiId: awssdk.String("blghshbgte")},
	}

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayV2Repository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway v2 models",
			mocks: func(repo *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllApis").Return(apis, nil)
				repo.On("ListAllApiModels", *apis[0].ApiId).Return([]*apigatewayv2.Model{}, nil).Once()
				repo.On("ListAllApiModels", *apis[1].ApiId).Return([]*apigatewayv2.Model{}, nil).Once()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple api gateway v2 models",
			mocks: func(repo *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllApis").Return(apis, nil)
				repo.On("ListAllApiModels", *apis[0].ApiId).Return([]*apigatewayv2.Model{
					{
						ModelId: awssdk.String("vdw6up"),
						Name:    awssdk.String("model1"),
					},
				}, nil).Once()
				repo.On("ListAllApiModels", *apis[1].ApiId).Return([]*apigatewayv2.Model{
					{
						ModelId: awssdk.String("bwhebj"),
						Name:    awssdk.String("model2"),
					},
				}, nil).Once()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "vdw6up")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayV2ModelResourceType)
				assert.Equal(t, "model1", *got[0].Attributes().GetString("name"))

				assert.Equal(t, got[1].ResourceId(), "bwhebj")
				assert.Equal(t, got[1].ResourceType(), resourceaws.AwsApiGatewayV2ModelResourceType)
				assert.Equal(t, "model2", *got[1].Attributes().GetString("name"))

			},
		},
		{
			test: "cannot list apis",
			mocks: func(repo *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2ModelResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2ModelResourceType, resourceaws.AwsApiGatewayV2ApiResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2ModelResourceType, resourceaws.AwsApiGatewayV2ApiResourceType),
		},
		{
			test: "cannot list api gateway v2 model",
			mocks: func(repo *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repo.On("ListAllApis").Return(apis, nil)
				repo.On("ListAllApiModels", *apis[0].ApiId).Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2ModelResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2ModelResourceType, resourceaws.AwsApiGatewayV2ModelResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayV2ModelResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayV2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayV2Repository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayV2ModelEnumerator(repo, factory))

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

func TestApiGatewayV2Stage(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayV2Repository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway v2 api",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return([]*apigatewayv2.Api{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "single api gateway v2 api with a single stage",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return([]*apigatewayv2.Api{
					{ApiId: awssdk.String("a-gateway")},
				}, nil)
				repository.On("ListAllApiStages", "a-gateway").
					Return([]*apigatewayv2.Stage{{
						StageName: awssdk.String("a-stage"),
					}}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)

				assert.Equal(t, "a-stage", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsApiGatewayV2StageResourceType, got[0].ResourceType())
			},
		},
		{
			test: "cannot list api gateway v2 apis",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2StageResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2ApiResourceType, resourceaws.AwsApiGatewayV2StageResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2StageResourceType, resourceaws.AwsApiGatewayV2ApiResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayV2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayV2Repository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayV2StageEnumerator(repo, factory))

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

func TestApiGatewayV2RouteResponse(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayV2Repository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway v2 route responses",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return([]*apigatewayv2.Api{
					{ApiId: awssdk.String("77ooqulkke")},
				}, nil)
				repository.On("ListAllApiRoutes", awssdk.String("77ooqulkke")).
					Return([]*apigatewayv2.Route{
						{RouteId: awssdk.String("liqc5u4")},
					}, nil)
				repository.On("ListAllApiRouteResponses", "77ooqulkke", "liqc5u4").
					Return([]*apigatewayv2.RouteResponse{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "single api gateway v2 route with one route response",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return([]*apigatewayv2.Api{
					{ApiId: awssdk.String("77ooqulkke")},
				}, nil)
				repository.On("ListAllApiRoutes", awssdk.String("77ooqulkke")).
					Return([]*apigatewayv2.Route{
						{RouteId: awssdk.String("liqc5u4")},
					}, nil)
				repository.On("ListAllApiRouteResponses", "77ooqulkke", "liqc5u4").
					Return([]*apigatewayv2.RouteResponse{
						{RouteResponseId: awssdk.String("nbw7vw")},
					}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)

				assert.Equal(t, got[0].ResourceId(), "nbw7vw")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayV2RouteResponseResourceType)
			},
		},
		{
			test: "cannot list api gateway v2 apis",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2RouteResponseResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2ApiResourceType, resourceaws.AwsApiGatewayV2RouteResponseResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2RouteResponseResourceType, resourceaws.AwsApiGatewayV2ApiResourceType),
		},
		{
			test: "cannot list api gateway v2 routes",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return([]*apigatewayv2.Api{
					{ApiId: awssdk.String("77ooqulkke")},
				}, nil)
				repository.On("ListAllApiRoutes", awssdk.String("77ooqulkke")).Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2RouteResponseResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2RouteResourceType, resourceaws.AwsApiGatewayV2RouteResponseResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2RouteResponseResourceType, resourceaws.AwsApiGatewayV2RouteResourceType),
		},
		{
			test: "cannot list api gateway v2 route responses",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return([]*apigatewayv2.Api{
					{ApiId: awssdk.String("77ooqulkke")},
				}, nil)
				repository.On("ListAllApiRoutes", awssdk.String("77ooqulkke")).
					Return([]*apigatewayv2.Route{
						{RouteId: awssdk.String("liqc5u4")},
					}, nil)
				repository.On("ListAllApiRouteResponses", "77ooqulkke", "liqc5u4").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2RouteResponseResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2RouteResponseResourceType, resourceaws.AwsApiGatewayV2RouteResponseResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2RouteResponseResourceType, resourceaws.AwsApiGatewayV2RouteResponseResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayV2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayV2Repository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayV2RouteResponseEnumerator(repo, factory))

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

func TestApiGatewayV2Mapping(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *repository.MockApiGatewayV2Repository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway v2 domains",
			mocks: func(repositoryV1 *repository.MockApiGatewayRepository, repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repositoryV1.On("ListAllDomainNames").Return([]*apigateway.DomainName{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "single api gateway v2 domain with a single mapping",
			mocks: func(repositoryV1 *repository.MockApiGatewayRepository, repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repositoryV1.On("ListAllDomainNames").Return([]*apigateway.DomainName{
					{DomainName: awssdk.String("example.com")},
				}, nil)
				repository.On("ListAllApiMappings", "example.com").
					Return([]*apigatewayv2.ApiMapping{{
						Stage:        awssdk.String("a-stage"),
						ApiId:        awssdk.String("foobar"),
						ApiMappingId: awssdk.String("barfoo"),
					}}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)

				assert.Equal(t, "barfoo", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsApiGatewayV2MappingResourceType, got[0].ResourceType())
			},
		},
		{
			test: "cannot list api gateway v2 domains",
			mocks: func(repositoryV1 *repository.MockApiGatewayRepository, repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repositoryV1.On("ListAllDomainNames").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2MappingResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayDomainNameResourceType, resourceaws.AwsApiGatewayV2MappingResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2MappingResourceType, resourceaws.AwsApiGatewayDomainNameResourceType),
		},
		{
			test: "cannot list api gateway v2 mappings",
			mocks: func(repositoryV1 *repository.MockApiGatewayRepository, repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repositoryV1.On("ListAllDomainNames").Return([]*apigateway.DomainName{
					{DomainName: awssdk.String("example.com")},
				}, nil)
				repository.On("ListAllApiMappings", "example.com").
					Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2MappingResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2MappingResourceType, resourceaws.AwsApiGatewayV2MappingResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayV2MappingResourceType),
		},
		{
			test: "returning mapping with invalid attributes",
			mocks: func(repositoryV1 *repository.MockApiGatewayRepository, repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repositoryV1.On("ListAllDomainNames").Return([]*apigateway.DomainName{
					{DomainName: awssdk.String("example.com")},
				}, nil)
				repository.On("ListAllApiMappings", "example.com").
					Return([]*apigatewayv2.ApiMapping{
						{
							ApiMappingId: awssdk.String("barfoo"),
						},
						{
							Stage:        awssdk.String("a-stage"),
							ApiId:        awssdk.String("foobar"),
							ApiMappingId: awssdk.String("foobar"),
						},
					}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, "barfoo", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsApiGatewayV2MappingResourceType, got[0].ResourceType())

				assert.Equal(t, "foobar", got[1].ResourceId())
				assert.Equal(t, resourceaws.AwsApiGatewayV2MappingResourceType, got[1].ResourceType())
			},
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepoV1 := &repository.MockApiGatewayRepository{}
			fakeRepo := &repository.MockApiGatewayV2Repository{}
			c.mocks(fakeRepoV1, fakeRepo, alerter)

			remoteLibrary.AddEnumerator(aws.NewApiGatewayV2MappingEnumerator(fakeRepo, fakeRepoV1, factory))

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
			fakeRepoV1.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestApiGatewayV2DomainName(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway v2 domain names",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllDomainNames").Return([]*apigateway.DomainName{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "single api gateway v2 domain name",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllDomainNames").Return([]*apigateway.DomainName{
					{DomainName: awssdk.String("b8r351.example.com")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)

				assert.Equal(t, got[0].ResourceId(), "b8r351.example.com")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayV2DomainNameResourceType)
			},
		},
		{
			test: "cannot list api gateway v2 domain names",
			mocks: func(repository *repository.MockApiGatewayRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllDomainNames").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2DomainNameResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2DomainNameResourceType, resourceaws.AwsApiGatewayV2DomainNameResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceaws.AwsApiGatewayV2DomainNameResourceType),
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

			remoteLibrary.AddEnumerator(aws.NewApiGatewayV2DomainNameEnumerator(repo, factory))

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

func TestApiGatewayV2IntegrationResponse(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockApiGatewayV2Repository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no api gateway v2 integration responses",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return([]*apigatewayv2.Api{
					{ApiId: awssdk.String("yw28nwdf34")},
				}, nil)
				repository.On("ListAllApiIntegrations", "yw28nwdf34").
					Return([]*apigatewayv2.Integration{
						{IntegrationId: awssdk.String("fmezvlh")},
					}, nil)
				repository.On("ListAllApiIntegrationResponses", "yw28nwdf34", "fmezvlh").
					Return([]*apigatewayv2.IntegrationResponse{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "single api gateway v2 integration with one integration response",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return([]*apigatewayv2.Api{
					{ApiId: awssdk.String("yw28nwdf34")},
				}, nil)
				repository.On("ListAllApiIntegrations", "yw28nwdf34").
					Return([]*apigatewayv2.Integration{
						{IntegrationId: awssdk.String("fmezvlh")},
					}, nil)
				repository.On("ListAllApiIntegrationResponses", "yw28nwdf34", "fmezvlh").
					Return([]*apigatewayv2.IntegrationResponse{
						{IntegrationResponseId: awssdk.String("sf67ti7")},
					}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)

				assert.Equal(t, got[0].ResourceId(), "sf67ti7")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsApiGatewayV2IntegrationResponseResourceType)
			},
		},
		{
			test: "cannot list api gateway v2 apis",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2IntegrationResponseResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2ApiResourceType, resourceaws.AwsApiGatewayV2IntegrationResponseResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2IntegrationResponseResourceType, resourceaws.AwsApiGatewayV2ApiResourceType),
		},
		{
			test: "cannot list api gateway v2 integrations",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return([]*apigatewayv2.Api{
					{ApiId: awssdk.String("yw28nwdf34")},
				}, nil)
				repository.On("ListAllApiIntegrations", "yw28nwdf34").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2IntegrationResponseResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2IntegrationResourceType, resourceaws.AwsApiGatewayV2IntegrationResponseResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2IntegrationResponseResourceType, resourceaws.AwsApiGatewayV2IntegrationResourceType),
		},
		{
			test: "cannot list api gateway v2 integration responses",
			mocks: func(repository *repository.MockApiGatewayV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllApis").Return([]*apigatewayv2.Api{
					{ApiId: awssdk.String("yw28nwdf34")},
				}, nil)
				repository.On("ListAllApiIntegrations", "yw28nwdf34").
					Return([]*apigatewayv2.Integration{
						{IntegrationId: awssdk.String("fmezvlh")},
					}, nil)
				repository.On("ListAllApiIntegrationResponses", "yw28nwdf34", "fmezvlh").Return(nil, dummyError)
				alerter.On("SendAlert", resourceaws.AwsApiGatewayV2IntegrationResponseResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2IntegrationResponseResourceType, resourceaws.AwsApiGatewayV2IntegrationResponseResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsApiGatewayV2IntegrationResponseResourceType, resourceaws.AwsApiGatewayV2IntegrationResponseResourceType),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockApiGatewayV2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ApiGatewayV2Repository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewApiGatewayV2IntegrationResponseEnumerator(repo, factory))

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
