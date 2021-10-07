package repository

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/postgresql/armpostgresql"
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm/common"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type PostgresqlRespository interface {
	ListAllServers() ([]*armpostgresql.Server, error)
	ListAllDatabasesByServer(resGroup string, serverName string) ([]*armpostgresql.Database, error)
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

type postgresqlDatabaseClientImpl struct {
	client *armpostgresql.DatabasesClient
}

type postgresqlDatabaseClient interface {
	ListByServer(context.Context, string, string, *armpostgresql.DatabasesListByServerOptions) (armpostgresql.DatabasesListByServerResponse, error)
}

func (c postgresqlDatabaseClientImpl) ListByServer(ctx context.Context, resGroup string, serverName string, options *armpostgresql.DatabasesListByServerOptions) (armpostgresql.DatabasesListByServerResponse, error) {
	return c.client.ListByServer(ctx, resGroup, serverName, options)
}

type postgresqlRepository struct {
	serversClient  postgresqlServersClient
	databaseClient postgresqlDatabaseClient
	cache          cache.Cache
}

func NewPostgresqlRepository(con *arm.Connection, config common.AzureProviderConfig, cache cache.Cache) *postgresqlRepository {
	return &postgresqlRepository{
		postgresqlServersClientImpl{client: armpostgresql.NewServersClient(con, config.SubscriptionID)},
		postgresqlDatabaseClientImpl{client: armpostgresql.NewDatabasesClient(con, config.SubscriptionID)},
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

func (s *postgresqlRepository) ListAllDatabasesByServer(resGroup string, serverName string) ([]*armpostgresql.Database, error) {
	cacheKey := fmt.Sprintf("postgresqlListAllDatabases_%s_%s", resGroup, serverName)
	if v := s.cache.Get(cacheKey); v != nil {
		return v.([]*armpostgresql.Database), nil
	}

	res, err := s.databaseClient.ListByServer(context.Background(), resGroup, serverName, &armpostgresql.DatabasesListByServerOptions{})
	if err != nil {
		return nil, err
	}

	s.cache.Put(cacheKey, res.Value)
	return res.Value, nil
}
