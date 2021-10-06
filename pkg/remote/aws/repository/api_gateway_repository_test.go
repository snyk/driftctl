package repository

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	awstest "github.com/cloudskiff/driftctl/test/aws"

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

				store.On("Get", "apigatewayListAllRestApis").Return(nil).Times(1)
				store.On("Put", "apigatewayListAllRestApis", apis).Return(false).Times(1)
			},
			want: apis,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				store.On("Get", "apigatewayListAllRestApis").Return(apis).Times(1)
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
	apis := []*apigateway.RestApi{
		{Id: aws.String("restapi1")},
		{Id: aws.String("restapi2")},
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
					}).Return(&apigateway.GetAuthorizersOutput{Items: apiAuthorizers[:2]}, nil).Once()

				client.On("GetAuthorizers",
					&apigateway.GetAuthorizersInput{
						RestApiId: aws.String("restapi2"),
					}).Return(&apigateway.GetAuthorizersOutput{Items: apiAuthorizers[2:]}, nil).Once()

				store.On("Get", "apigatewayListAllRestApiAuthorizers_api_restapi1").Return(nil).Times(1)
				store.On("Put", "apigatewayListAllRestApiAuthorizers_api_restapi1", apiAuthorizers[:2]).Return(false).Times(1)
				store.On("Get", "apigatewayListAllRestApiAuthorizers_api_restapi2").Return(nil).Times(1)
				store.On("Put", "apigatewayListAllRestApiAuthorizers_api_restapi2", apiAuthorizers[2:]).Return(false).Times(1)
			},
			want: apiAuthorizers,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				store.On("Get", "apigatewayListAllRestApiAuthorizers_api_restapi1").Return(apiAuthorizers[:2]).Times(1)
				store.On("Get", "apigatewayListAllRestApiAuthorizers_api_restapi2").Return(apiAuthorizers[2:]).Times(1)
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
			got, err := r.ListAllRestApiAuthorizers(apis)
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

				store.On("Get", "apigatewayListAllRestApiStages_api_restapi1").Return(nil).Times(1)
				store.On("Put", "apigatewayListAllRestApiStages_api_restapi1", apiStages).Return(false).Times(1)
			},
			want: apiStages,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				store.On("Get", "apigatewayListAllRestApiStages_api_restapi1").Return(apiStages).Times(1)
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
					},
					mock.MatchedBy(func(callback func(res *apigateway.GetResourcesOutput, lastPage bool) bool) bool {
						callback(&apigateway.GetResourcesOutput{
							Items: apiResources,
						}, true)
						return true
					})).Return(nil).Once()

				store.On("Get", "apigatewayListAllRestApiResources_api_restapi1").Return(nil).Times(1)
				store.On("Put", "apigatewayListAllRestApiResources_api_restapi1", apiResources).Return(false).Times(1)
			},
			want: apiResources,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				store.On("Get", "apigatewayListAllRestApiResources_api_restapi1").Return(apiResources).Times(1)
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

				store.On("Get", "apigatewayListAllDomainNames").Return(nil).Times(1)
				store.On("Put", "apigatewayListAllDomainNames", domainNames).Return(false).Times(1)
			},
			want: domainNames,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeApiGateway, store *cache.MockCache) {
				store.On("Get", "apigatewayListAllDomainNames").Return(domainNames).Times(1)
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
