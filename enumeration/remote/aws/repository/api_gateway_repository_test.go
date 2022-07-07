package repository

import (
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/pkg/errors"
	awstest "github.com/snyk/driftctl/test/aws"

	"github.com/stretchr/testify/mock"

	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
)

func Test_apigatewayRepository_ListAllRestApis(t *testing.T) {
	apis := []*apigateway.RestApi{
		{Id: aws.String("restapi1")},
		{Id: aws.String("restapi2")},
		{Id: aws.String("restapi3")},
		{Id: aws.String("restapi4")},
		{Id: aws.String("restapi5")},
		{Id: aws.String("restapi6")},
	}

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeApiGateway, store *cache.MockCache)
		want    []*apigateway.RestApi
		wantErr error
	}{
		{
			name: "list multiple rest apis",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				client.On("GetRestApisPages",
					&apigateway.GetRestApisInput{},
					mock.MatchedBy(func(callback func(res *apigateway.GetRestApisOutput, lastPage bool) bool) bool {
						callback(&apigateway.GetRestApisOutput{
							Items: apis[:3],
						}, false)
						callback(&apigateway.GetRestApisOutput{
							Items: apis[3:],
						}, true)
						return true
					})).Return(nil).Once()

				store.On("GetAndLock", "apigatewayListAllRestApis").Return(nil).Times(1)
				store.On("Unlock", "apigatewayListAllRestApis").Times(1)
				store.On("Put", "apigatewayListAllRestApis", apis).Return(false).Times(1)
			},
			want: apis,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				store.On("GetAndLock", "apigatewayListAllRestApis").Return(apis).Times(1)
				store.On("Unlock", "apigatewayListAllRestApis").Times(1)
			},
			want: apis,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeApiGateway{}
			tt.mocks(client, store)
			r := &apigatewayRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllRestApis()
			assert.Equal(t, tt.wantErr, err)

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			store.AssertExpectations(t)
			client.AssertExpectations(t)
		})
	}
}

func Test_apigatewayRepository_GetAccount(t *testing.T) {
	account := &apigateway.Account{
		CloudwatchRoleArn: aws.String("arn:aws:iam::017011014111:role/api_gateway_cloudwatch_global"),
	}

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeApiGateway, store *cache.MockCache)
		want    *apigateway.Account
		wantErr error
	}{
		{
			name: "get a single account",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				client.On("GetAccount", &apigateway.GetAccountInput{}).Return(account, nil).Once()

				store.On("Get", "apigatewayGetAccount").Return(nil).Times(1)
				store.On("Put", "apigatewayGetAccount", account).Return(false).Times(1)
			},
			want: account,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				store.On("Get", "apigatewayGetAccount").Return(account).Times(1)
			},
			want: account,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeApiGateway{}
			tt.mocks(client, store)
			r := &apigatewayRepository{
				client: client,
				cache:  store,
			}
			got, err := r.GetAccount()
			assert.Equal(t, tt.wantErr, err)

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			store.AssertExpectations(t)
			client.AssertExpectations(t)
		})
	}
}

func Test_apigatewayRepository_ListAllApiKeys(t *testing.T) {
	keys := []*apigateway.ApiKey{
		{Id: aws.String("apikey1")},
		{Id: aws.String("apikey2")},
		{Id: aws.String("apikey3")},
		{Id: aws.String("apikey4")},
		{Id: aws.String("apikey5")},
		{Id: aws.String("apikey6")},
	}

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeApiGateway, store *cache.MockCache)
		want    []*apigateway.ApiKey
		wantErr error
	}{
		{
			name: "list multiple api keys",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				client.On("GetApiKeysPages",
					&apigateway.GetApiKeysInput{},
					mock.MatchedBy(func(callback func(res *apigateway.GetApiKeysOutput, lastPage bool) bool) bool {
						callback(&apigateway.GetApiKeysOutput{
							Items: keys[:3],
						}, false)
						callback(&apigateway.GetApiKeysOutput{
							Items: keys[3:],
						}, true)
						return true
					})).Return(nil).Once()

				store.On("Get", "apigatewayListAllApiKeys").Return(nil).Times(1)
				store.On("Put", "apigatewayListAllApiKeys", keys).Return(false).Times(1)
			},
			want: keys,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				store.On("Get", "apigatewayListAllApiKeys").Return(keys).Times(1)
			},
			want: keys,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeApiGateway{}
			tt.mocks(client, store)
			r := &apigatewayRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllApiKeys()
			assert.Equal(t, tt.wantErr, err)

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			store.AssertExpectations(t)
			client.AssertExpectations(t)
		})
	}
}

func Test_apigatewayRepository_ListAllRestApiAuthorizers(t *testing.T) {
	api := &apigateway.RestApi{
		Id: aws.String("restapi1"),
	}

	apiAuthorizers := []*apigateway.Authorizer{
		{Id: aws.String("resource1")},
		{Id: aws.String("resource2")},
		{Id: aws.String("resource3")},
		{Id: aws.String("resource4")},
	}

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeApiGateway, store *cache.MockCache)
		want    []*apigateway.Authorizer
		wantErr error
	}{
		{
			name: "list multiple rest api authorizers",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				client.On("GetAuthorizers",
					&apigateway.GetAuthorizersInput{
						RestApiId: aws.String("restapi1"),
					}).Return(&apigateway.GetAuthorizersOutput{Items: apiAuthorizers}, nil).Once()

				store.On("Get", "apigatewayListAllRestApiAuthorizers_api_restapi1").Return(nil).Times(1)
				store.On("Put", "apigatewayListAllRestApiAuthorizers_api_restapi1", apiAuthorizers).Return(false).Times(1)
			},
			want: apiAuthorizers,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				store.On("Get", "apigatewayListAllRestApiAuthorizers_api_restapi1").Return(apiAuthorizers).Times(1)
			},
			want: apiAuthorizers,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeApiGateway{}
			tt.mocks(client, store)
			r := &apigatewayRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllRestApiAuthorizers(*api.Id)
			assert.Equal(t, tt.wantErr, err)

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			store.AssertExpectations(t)
			client.AssertExpectations(t)
		})
	}
}

func Test_apigatewayRepository_ListAllRestApiStages(t *testing.T) {
	api := &apigateway.RestApi{
		Id: aws.String("restapi1"),
	}

	apiStages := []*apigateway.Stage{
		{StageName: aws.String("stage1")},
		{StageName: aws.String("stage2")},
		{StageName: aws.String("stage3")},
		{StageName: aws.String("stage4")},
	}

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeApiGateway, store *cache.MockCache)
		want    []*apigateway.Stage
		wantErr error
	}{
		{
			name: "list multiple rest api stages",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				client.On("GetStages",
					&apigateway.GetStagesInput{
						RestApiId: aws.String("restapi1"),
					}).Return(&apigateway.GetStagesOutput{Item: apiStages}, nil).Once()

				store.On("GetAndLock", "apigatewayListAllRestApiStages_api_restapi1").Return(nil).Times(1)
				store.On("Unlock", "apigatewayListAllRestApiStages_api_restapi1").Times(1)
				store.On("Put", "apigatewayListAllRestApiStages_api_restapi1", apiStages).Return(false).Times(1)
			},
			want: apiStages,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				store.On("GetAndLock", "apigatewayListAllRestApiStages_api_restapi1").Return(apiStages).Times(1)
				store.On("Unlock", "apigatewayListAllRestApiStages_api_restapi1").Times(1)
			},
			want: apiStages,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeApiGateway{}
			tt.mocks(client, store)
			r := &apigatewayRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllRestApiStages(*api.Id)
			assert.Equal(t, tt.wantErr, err)

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			store.AssertExpectations(t)
			client.AssertExpectations(t)
		})
	}
}

func Test_apigatewayRepository_ListAllRestApiResources(t *testing.T) {
	api := &apigateway.RestApi{
		Id: aws.String("restapi1"),
	}

	apiResources := []*apigateway.Resource{
		{Id: aws.String("resource1")},
		{Id: aws.String("resource2")},
		{Id: aws.String("resource3")},
		{Id: aws.String("resource4")},
	}

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeApiGateway, store *cache.MockCache)
		want    []*apigateway.Resource
		wantErr error
	}{
		{
			name: "list multiple rest api resources",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				client.On("GetResourcesPages",
					&apigateway.GetResourcesInput{
						RestApiId: aws.String("restapi1"),
						Embed:     []*string{aws.String("methods")},
					},
					mock.MatchedBy(func(callback func(res *apigateway.GetResourcesOutput, lastPage bool) bool) bool {
						callback(&apigateway.GetResourcesOutput{
							Items: apiResources,
						}, true)
						return true
					})).Return(nil).Once()

				store.On("GetAndLock", "apigatewayListAllRestApiResources_api_restapi1").Return(nil).Times(1)
				store.On("Unlock", "apigatewayListAllRestApiResources_api_restapi1").Times(1)
				store.On("Put", "apigatewayListAllRestApiResources_api_restapi1", apiResources).Return(false).Times(1)
			},
			want: apiResources,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				store.On("GetAndLock", "apigatewayListAllRestApiResources_api_restapi1").Return(apiResources).Times(1)
				store.On("Unlock", "apigatewayListAllRestApiResources_api_restapi1").Times(1)
			},
			want: apiResources,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeApiGateway{}
			tt.mocks(client, store)
			r := &apigatewayRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllRestApiResources(*api.Id)
			assert.Equal(t, tt.wantErr, err)

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			store.AssertExpectations(t)
			client.AssertExpectations(t)
		})
	}
}

func Test_apigatewayRepository_ListAllDomainNames(t *testing.T) {
	domainNames := []*apigateway.DomainName{
		{DomainName: aws.String("domainName1")},
		{DomainName: aws.String("domainName2")},
		{DomainName: aws.String("domainName3")},
		{DomainName: aws.String("domainName4")},
		{DomainName: aws.String("domainName5")},
		{DomainName: aws.String("domainName6")},
	}

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeApiGateway, store *cache.MockCache)
		want    []*apigateway.DomainName
		wantErr error
	}{
		{
			name: "list multiple domain names",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				client.On("GetDomainNamesPages",
					&apigateway.GetDomainNamesInput{},
					mock.MatchedBy(func(callback func(res *apigateway.GetDomainNamesOutput, lastPage bool) bool) bool {
						callback(&apigateway.GetDomainNamesOutput{
							Items: domainNames[:3],
						}, false)
						callback(&apigateway.GetDomainNamesOutput{
							Items: domainNames[3:],
						}, true)
						return true
					})).Return(nil).Once()

				store.On("GetAndLock", "apigatewayListAllDomainNames").Return(nil).Times(1)
				store.On("Unlock", "apigatewayListAllDomainNames").Times(1)
				store.On("Put", "apigatewayListAllDomainNames", domainNames).Return(false).Times(1)
			},
			want: domainNames,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				store.On("GetAndLock", "apigatewayListAllDomainNames").Return(domainNames).Times(1)
				store.On("Unlock", "apigatewayListAllDomainNames").Times(1)
			},
			want: domainNames,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeApiGateway{}
			tt.mocks(client, store)
			r := &apigatewayRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllDomainNames()
			assert.Equal(t, tt.wantErr, err)

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			store.AssertExpectations(t)
			client.AssertExpectations(t)
		})
	}
}

func Test_apigatewayRepository_ListAllVpcLinks(t *testing.T) {
	vpcLinks := []*apigateway.UpdateVpcLinkOutput{
		{Id: aws.String("vpcLink1")},
		{Id: aws.String("vpcLink2")},
		{Id: aws.String("vpcLink3")},
		{Id: aws.String("vpcLink4")},
		{Id: aws.String("vpcLink5")},
		{Id: aws.String("vpcLink6")},
	}

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeApiGateway, store *cache.MockCache)
		want    []*apigateway.UpdateVpcLinkOutput
		wantErr error
	}{
		{
			name: "list multiple vpc links",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				client.On("GetVpcLinksPages",
					&apigateway.GetVpcLinksInput{},
					mock.MatchedBy(func(callback func(res *apigateway.GetVpcLinksOutput, lastPage bool) bool) bool {
						callback(&apigateway.GetVpcLinksOutput{
							Items: vpcLinks[:3],
						}, false)
						callback(&apigateway.GetVpcLinksOutput{
							Items: vpcLinks[3:],
						}, true)
						return true
					})).Return(nil).Once()

				store.On("Get", "apigatewayListAllVpcLinks").Return(nil).Times(1)
				store.On("Put", "apigatewayListAllVpcLinks", vpcLinks).Return(false).Times(1)
			},
			want: vpcLinks,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				store.On("Get", "apigatewayListAllVpcLinks").Return(vpcLinks).Times(1)
			},
			want: vpcLinks,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeApiGateway{}
			tt.mocks(client, store)
			r := &apigatewayRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllVpcLinks()
			assert.Equal(t, tt.wantErr, err)

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			store.AssertExpectations(t)
			client.AssertExpectations(t)
		})
	}
}

func Test_apigatewayRepository_ListAllRestApiRequestValidators(t *testing.T) {
	api := &apigateway.RestApi{
		Id: aws.String("restapi1"),
	}

	requestValidators := []*apigateway.UpdateRequestValidatorOutput{
		{Id: aws.String("reqVal1")},
		{Id: aws.String("reqVal2")},
		{Id: aws.String("reqVal3")},
		{Id: aws.String("reqVal4")},
	}

	remoteError := errors.New("remote error")

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeApiGateway, store *cache.MockCache)
		want    []*apigateway.UpdateRequestValidatorOutput
		wantErr error
	}{
		{
			name: "list multiple rest api request validators",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				client.On("GetRequestValidators",
					&apigateway.GetRequestValidatorsInput{
						RestApiId: aws.String("restapi1"),
					}).Return(&apigateway.GetRequestValidatorsOutput{Items: requestValidators}, nil).Once()

				store.On("Get", "apigatewayListAllRestApiRequestValidators_api_restapi1").Return(nil).Times(1)
				store.On("Put", "apigatewayListAllRestApiRequestValidators_api_restapi1", requestValidators).Return(false).Times(1)
			},
			want: requestValidators,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				store.On("Get", "apigatewayListAllRestApiRequestValidators_api_restapi1").Return(requestValidators).Times(1)
			},
			want: requestValidators,
		},
		{
			name: "should return remote error",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				client.On("GetRequestValidators",
					&apigateway.GetRequestValidatorsInput{
						RestApiId: aws.String("restapi1"),
					}).Return(nil, remoteError).Once()

				store.On("Get", "apigatewayListAllRestApiRequestValidators_api_restapi1").Return(nil).Times(1)
			},
			wantErr: remoteError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeApiGateway{}
			tt.mocks(client, store)
			r := &apigatewayRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllRestApiRequestValidators(*api.Id)
			assert.Equal(t, tt.wantErr, err)

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			store.AssertExpectations(t)
			client.AssertExpectations(t)
		})
	}
}

func Test_apigatewayRepository_ListAllDomainNameBasePathMappings(t *testing.T) {
	domainName := &apigateway.DomainName{
		DomainName: aws.String("domainName1"),
	}

	mappings := []*apigateway.BasePathMapping{
		{BasePath: aws.String("path1")},
		{BasePath: aws.String("path2")},
		{BasePath: aws.String("path3")},
		{BasePath: aws.String("path4")},
	}

	remoteError := errors.New("remote error")

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeApiGateway, store *cache.MockCache)
		want    []*apigateway.BasePathMapping
		wantErr error
	}{
		{
			name: "list multiple domain name base path mappings",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				client.On("GetBasePathMappingsPages",
					&apigateway.GetBasePathMappingsInput{
						DomainName: aws.String("domainName1"),
					},
					mock.MatchedBy(func(callback func(res *apigateway.GetBasePathMappingsOutput, lastPage bool) bool) bool {
						callback(&apigateway.GetBasePathMappingsOutput{
							Items: mappings,
						}, true)
						return true
					})).Return(nil).Once()

				store.On("Get", "apigatewayListAllDomainNameBasePathMappings_domainName_domainName1").Return(nil).Times(1)
				store.On("Put", "apigatewayListAllDomainNameBasePathMappings_domainName_domainName1", mappings).Return(false).Times(1)
			},
			want: mappings,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				store.On("Get", "apigatewayListAllDomainNameBasePathMappings_domainName_domainName1").Return(mappings).Times(1)
			},
			want: mappings,
		},
		{
			name: "should return remote error",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				client.On("GetBasePathMappingsPages",
					&apigateway.GetBasePathMappingsInput{
						DomainName: aws.String("domainName1"),
					}, mock.AnythingOfType("func(*apigateway.GetBasePathMappingsOutput, bool) bool")).Return(remoteError).Once()

				store.On("Get", "apigatewayListAllDomainNameBasePathMappings_domainName_domainName1").Return(nil).Times(1)
			},
			wantErr: remoteError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeApiGateway{}
			tt.mocks(client, store)
			r := &apigatewayRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllDomainNameBasePathMappings(*domainName.DomainName)
			assert.Equal(t, tt.wantErr, err)

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			store.AssertExpectations(t)
			client.AssertExpectations(t)
		})
	}
}

func Test_apigatewayRepository_ListAllRestApiModels(t *testing.T) {
	api := &apigateway.RestApi{
		Id: aws.String("restapi1"),
	}

	apiModels := []*apigateway.Model{
		{Id: aws.String("model1")},
		{Id: aws.String("model2")},
		{Id: aws.String("model3")},
		{Id: aws.String("model4")},
	}

	remoteError := errors.New("remote error")

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeApiGateway, store *cache.MockCache)
		want    []*apigateway.Model
		wantErr error
	}{
		{
			name: "list multiple rest api models",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				client.On("GetModelsPages",
					&apigateway.GetModelsInput{
						RestApiId: aws.String("restapi1"),
					},
					mock.MatchedBy(func(callback func(res *apigateway.GetModelsOutput, lastPage bool) bool) bool {
						callback(&apigateway.GetModelsOutput{
							Items: apiModels,
						}, true)
						return true
					})).Return(nil).Once()

				store.On("Get", "apigatewayListAllRestApiModels_api_restapi1").Return(nil).Times(1)
				store.On("Put", "apigatewayListAllRestApiModels_api_restapi1", apiModels).Return(false).Times(1)
			},
			want: apiModels,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				store.On("Get", "apigatewayListAllRestApiModels_api_restapi1").Return(apiModels).Times(1)
			},
			want: apiModels,
		},
		{
			name: "should return remote error",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				client.On("GetModelsPages",
					&apigateway.GetModelsInput{
						RestApiId: aws.String("restapi1"),
					}, mock.AnythingOfType("func(*apigateway.GetModelsOutput, bool) bool")).Return(remoteError).Once()

				store.On("Get", "apigatewayListAllRestApiModels_api_restapi1").Return(nil).Times(1)
			},
			wantErr: remoteError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeApiGateway{}
			tt.mocks(client, store)
			r := &apigatewayRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllRestApiModels(*api.Id)
			assert.Equal(t, tt.wantErr, err)

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			store.AssertExpectations(t)
			client.AssertExpectations(t)
		})
	}
}

func Test_apigatewayRepository_ListAllRestApiGatewayResponses(t *testing.T) {
	api := &apigateway.RestApi{
		Id: aws.String("restapi1"),
	}

	gtwResponses := []*apigateway.UpdateGatewayResponseOutput{
		{ResponseType: aws.String("ACCESS_DENIED")},
		{ResponseType: aws.String("DEFAULT_4XX")},
		{ResponseType: aws.String("DEFAULT_5XX")},
		{ResponseType: aws.String("UNAUTHORIZED")},
	}

	remoteError := errors.New("remote error")

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeApiGateway, store *cache.MockCache)
		want    []*apigateway.UpdateGatewayResponseOutput
		wantErr error
	}{
		{
			name: "list multiple rest api gateway responses",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				client.On("GetGatewayResponses",
					&apigateway.GetGatewayResponsesInput{
						RestApiId: aws.String("restapi1"),
					}).Return(&apigateway.GetGatewayResponsesOutput{Items: gtwResponses}, nil).Once()

				store.On("Get", "apigatewayListAllRestApiGatewayResponses_api_restapi1").Return(nil).Times(1)
				store.On("Put", "apigatewayListAllRestApiGatewayResponses_api_restapi1", gtwResponses).Return(false).Times(1)
			},
			want: gtwResponses,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				store.On("Get", "apigatewayListAllRestApiGatewayResponses_api_restapi1").Return(gtwResponses).Times(1)
			},
			want: gtwResponses,
		},
		{
			name: "should return remote error",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				client.On("GetGatewayResponses",
					&apigateway.GetGatewayResponsesInput{
						RestApiId: aws.String("restapi1"),
					}).Return(nil, remoteError).Once()

				store.On("Get", "apigatewayListAllRestApiGatewayResponses_api_restapi1").Return(nil).Times(1)
			},
			wantErr: remoteError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeApiGateway{}
			tt.mocks(client, store)
			r := &apigatewayRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllRestApiGatewayResponses(*api.Id)
			assert.Equal(t, tt.wantErr, err)

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			store.AssertExpectations(t)
			client.AssertExpectations(t)
		})
	}
}
