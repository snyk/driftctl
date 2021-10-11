package repository

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/postgresql/armpostgresql"
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm/common"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type PostgresqlRespository interface {
	ListAllServers() ([]*armpostgresql.Server, error)
}

type postgresqlServersClientImpl struct {
	client *armpostgresql.ServersClient
}

type postgresqlServersClient interface {
	List(context.Context, *armpostgresql.ServersListOptions) (armpostgresql.ServersListResponse, error)
}

func (c postgresqlServersClientImpl) List(ctx context.Context, options *armpostgresql.ServersListOptions) (armpostgresql.ServersListResponse, error) {
	return c.client.List(ctx, options)
}

type postgresqlRepository struct {
	serversClient postgresqlServersClient
	cache         cache.Cache
}

func NewPostgresqlRepository(con *arm.Connection, config common.AzureProviderConfig, cache cache.Cache) *postgresqlRepository {
	return &postgresqlRepository{
		postgresqlServersClientImpl{client: armpostgresql.NewServersClient(con, config.SubscriptionID)},
		cache,
	}
}

func (s *postgresqlRepository) ListAllServers() ([]*armpostgresql.Server, error) {
	cacheKey := "postgresqlListAllServers"
	if v := s.cache.Get(cacheKey); v != nil {
		return v.([]*armpostgresql.Server), nil
	}

	res, err := s.serversClient.List(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	s.cache.Put(cacheKey, res.Value)
	return res.Value, nil
}
