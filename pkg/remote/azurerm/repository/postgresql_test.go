package repository

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/postgresql/armpostgresql"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Postgresql_ListAllServers(t *testing.T) {
	expectedResults := []*armpostgresql.Server{
		{
			TrackedResource: armpostgresql.TrackedResource{
				Resource: armpostgresql.Resource{
					ID: to.StringPtr("postgresql-server-1"),
				},
			},
		},
		{
			TrackedResource: armpostgresql.TrackedResource{
				Resource: armpostgresql.Resource{
					ID: to.StringPtr("postgresql-server-2"),
				},
			},
		},
	}

	testcases := []struct {
		name     string
		mocks    func(*mockPostgresqlServersClient, *cache.MockCache)
		expected []*armpostgresql.Server
		wantErr  string
	}{
		{
			name: "should return postgres servers",
			mocks: func(client *mockPostgresqlServersClient, mockCache *cache.MockCache) {
				client.On("List", context.Background(), mock.Anything).Return(armpostgresql.ServersListResponse{
					ServersListResult: armpostgresql.ServersListResult{
						ServerListResult: armpostgresql.ServerListResult{
							Value: expectedResults,
						},
					},
				}, nil).Times(1)

				mockCache.On("Get", "postgresqlListAllServers").Return(nil).Times(1)
				mockCache.On("Put", "postgresqlListAllServers", expectedResults).Return(false).Times(1)
			},
			expected: expectedResults,
		},
		{
			name: "should hit cache and return postgres servers",
			mocks: func(client *mockPostgresqlServersClient, mockCache *cache.MockCache) {
				mockCache.On("Get", "postgresqlListAllServers").Return(expectedResults).Times(1)
			},
			expected: expectedResults,
		},
		{
			name: "should return remote error",
			mocks: func(client *mockPostgresqlServersClient, mockCache *cache.MockCache) {
				client.On("List", context.Background(), mock.Anything).Return(armpostgresql.ServersListResponse{}, errors.New("remote error")).Times(1)

				mockCache.On("Get", "postgresqlListAllServers").Return(nil).Times(1)
			},
			wantErr: "remote error",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := &mockPostgresqlServersClient{}
			mockCache := &cache.MockCache{}

			tt.mocks(fakeClient, mockCache)

			s := &postgresqlRepository{
				serversClient: fakeClient,
				cache:         mockCache,
			}
			got, err := s.ListAllServers()
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			} else {
				assert.Nil(t, err)
			}

			fakeClient.AssertExpectations(t)
			mockCache.AssertExpectations(t)

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ListAllResourceGroups() got = %v, want %v", got, tt.expected)
			}
		})
	}
}
